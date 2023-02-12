package controller

import (
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/gin-gonic/gin"
	"strconv"
)

func Listening(gp *gin.RouterGroup) {
	gp.GET("", getListening)
	gp.GET("/:room_id", getListenRoom)
}

func getListenRoom(c *gin.Context) {

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
			c.IndentedJSON(404, gin.H{
				"error": "房間不存在",
			})
			return
		}

		c.IndentedJSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.IndentedJSON(200, room)
}

func getListening(c *gin.Context) {

	listens := blive.GetEntered()

	c.JSON(200, gin.H{
		"total_started_count":   len(blive.GetListening()),
		"excepted_count":        len(blive.GetExcepted()),
		"total_listening_count": len(listens),
		"rooms":                 listens,
	})
}
