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
	gp.GET("/:room_id")
}

func GetListenRoom(c *gin.Context) {

	id, err := strconv.ParseInt(c.Param(":room_id"), 10, 64)

	if err != nil {
		c.IndentedJSON(400, gin.H{
			"error": "room_id 必須為數字",
		})
		return
	}

	room, err := blive.GetLiveInfo(id)

	if err != nil {
		log.Warnf("嘗試獲取房間 %v 的直播資訊時出現錯誤: %v", id, err)
		c.IndentedJSON(500, gin.H{
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
