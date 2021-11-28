package subscribe

import (
	mapset "github.com/deckarep/golang-set"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"strings"
	"time"
)

func Register(gp *gin.RouterGroup) {
	gp.GET("", GetSubscriptions)
	gp.POST("", Subscribe)
	gp.DELETE("", DeleteSubscribe)
}

func GetSubscriptions(c *gin.Context) {
	list, ok := subscriber.Get(c.ClientIP())
	if !ok {
		list = []int64{}
	}
	c.IndentedJSON(200, list)
}

func DeleteSubscribe(c *gin.Context) {
	subscriber.Delete(c.ClientIP())
	c.Status(200)
}

func Subscribe(c *gin.Context) {
	subArr, ok := c.GetPostFormArray("subscribes")
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "缺少 `subscribes` 數值(訂閱列表)"})
		return
	}
	if len(subArr) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "訂閱列表不能為空"})
		return
	}
	log.Printf("%v is going to subscribe %v", c.ClientIP(), strings.Join(subArr, ", "))
	roomSet := mapset.NewSet()
	for _, arr := range subArr {

		roomId, err := strconv.ParseInt(arr, 10, 64)

		if err != nil {
			log.Println("Cannot parse room: ", err.Error())
			continue
		}

		if realRoom, roomErr := api.GetRealRoom(roomId); realRoom > 0 {
			roomSet.Add(realRoom)
		} else if roomErr != nil {
			_ = c.Error(roomErr)
			return
		}

	}

	// 如果之前尚未有過訂閱 (即新增而不是更新)
	if _, subBefore := subscriber.Get(c.ClientIP()); !subBefore {
		// 設置如果五分鐘後尚未連線 WebSocket 就清除訂閱記憶
		subscriber.ExpireAfter(c.ClientIP(), time.After(time.Minute*5))
	}

	// 有生之年我居然還要用 loop 轉泛型 array 同 type array, 醉了
	roomArr := roomSet.ToSlice()
	rooms := make([]int64, len(roomArr))

	for i, v := range roomArr {
		rooms[i] = v.(int64)
	}

	subscriber.Update(c.ClientIP(), rooms)
	c.IndentedJSON(200, rooms)
}
