package listening

import (
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strconv"
)

var log = logrus.WithField("controller", "listening")

func Register(gp *gin.RouterGroup) {
	gp.GET("", GetListening)
	gp.GET("/:room_id", GetListenRoom)
}

func GetListenRoom(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param("room_id"), 10, 64)

	if err != nil {
		c.IndentedJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	room, err := blive.GetListeningInfo(id)

	if err != nil {

		if err == blive.ErrNotFound {
			log.Infof("用戶索取 %v 房間資訊時不存在 (%v)", id, c.ClientIP())
			c.IndentedJSON(404, gin.H{
				"error": "房間不存在",
			})
			return
		}

		log.Warnf("嘗試獲取房間 %v 的直播資訊時出現錯誤: %v (%v)", id, err, c.ClientIP())
		c.IndentedJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.IndentedJSON(200, room)
}

func GetListening(c *gin.Context) {

	listens := blive.GetListening()

	c.JSON(200, gin.H{
		"total_listening_count": len(listens),
		"rooms":                 listens,
	})
}
