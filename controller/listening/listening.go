package listening

import (
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("controller", "listening")

func Register(gp *gin.RouterGroup) {
	gp.GET("", GetListening)
}

func GetListening(c *gin.Context) {

	listens := blive.GetListening()

	liveInfos := make([]*blive.LiveInfo, 0)

	for _, listen := range listens {
		roomId := listen.(int64)
		info, err := blive.GetLiveInfo(roomId)
		if err != nil {
			log.Warnf("獲取 %v 的房間資訊時出現錯誤: %v", roomId, err)
			continue
		}
		liveInfos = append(liveInfos, info)
	}

	c.JSON(200, gin.H{
		"total_listening_count": len(listens),
		"rooms":                 liveInfos,
	})
}
