package controller

import (
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("module", "controller")

func Subscribe(gp *gin.RouterGroup) {
	gp.GET("", getSubscribes)
	gp.POST("", updateSubscribe)
	gp.DELETE("", clearSubscribe)
	gp.PUT("add", addSubscribe)
	gp.PUT("remove", removeSubscribe)
}

func getSubscribes(c *gin.Context) {
	list, _ := subscriber.GetSubscribes(c.GetString("identifier"))
	c.IndentedJSON(200, list)
}

func clearSubscribe(c *gin.Context) {
	subscriber.RemoveSubscriber(c.GetString("identifier"))
	c.Status(200)
}

func addSubscribe(c *gin.Context) {

	dontCheck := c.Query("validate") == "false" // 是否不检查房间讯息
	rooms, ok := getSubscribesArr(c, !dontCheck)

	if !ok {
		return
	}

	id := c.GetString("identifier")

	logger.Infof("用戶 %v 新增訂閱 %v \n", id, rooms)

	sub := subscriber.GetSubscriber(id)
	sub.UpdateSubscribe(rooms)

	go subscriber.ActiveExpire(id, 30*time.Minute)

	c.IndentedJSON(200, sub.GetSubscribed())
}

func removeSubscribe(c *gin.Context) {

	rooms, ok := getSubscribesArr(c, false) // 刪除訂閱不檢查房間訊息是否存在

	if !ok {
		return
	}

	id := c.GetString("identifier")

	logger.Infof("用戶 %v 移除訂閱 %v \n", id, rooms)

	if !subscriber.HasSubscriber(id) {
		c.IndentedJSON(400, gin.H{"error": "刪除失敗，你尚未遞交過任何訂閱"})
		return
	}

	sub := subscriber.GetSubscriber(id)
	sub.RemoveSubscribes(rooms)

	c.IndentedJSON(200, sub.GetSubscribed())
}

func updateSubscribe(c *gin.Context) {
	dontCheck := c.Query("validate") == "false" // 是否不检查房间讯息
	rooms, ok := getSubscribesArr(c, !dontCheck)

	if !ok {
		return
	}

	id := c.GetString("identifier")

	logger.Infof("用戶 %v 設置訂閱 %v \n", id, rooms)

	subscriber.GetSubscriber(id).UpdateSubscribe(rooms)
	go subscriber.ActiveExpire(id, time.Minute*5)
	c.IndentedJSON(200, rooms)
}

func getSubscribesArr(c *gin.Context, checkExist bool) ([]int64, bool) {

	subArr, ok := c.GetPostFormArray("subscribes")
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "缺少 `subscribes` 數值(訂閱列表)"})
		return nil, false
	}
	if len(subArr) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "訂閱列表不能為空"})
		return nil, false
	}

	roomSet := mapset.NewSet[int64]()

	for _, arr := range subArr {

		roomId, err := strconv.ParseInt(arr, 10, 64)

		if err != nil {
			logger.Warn("cannot parse room: ", err.Error())
			continue
		}

		if checkExist {

			realRoom, roomErr := api.GetRealRoom(roomId)

			if roomErr != nil {
				logger.Warnf("獲取房間訊息時出現錯誤: %v", roomErr)
				_ = c.Error(roomErr)
				return nil, false
			} else {
				if realRoom > 0 {
					roomSet.Add(realRoom)
				} else {
					logger.Warnf("房間 %v 無效，已略過 \n", roomId)
				}
			}
		} else {
			roomSet.Add(roomId)
		}

	}

	return roomSet.ToSlice(), true
}
