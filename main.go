package main

import (
	"github.com/eric2788/biligo-live-ws/controller/subscribe"
	ws "github.com/eric2788/biligo-live-ws/controller/websocket"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	router := gin.Default()

	router.Use(CORS())
	router.Use(ErrorHandler)

	router.GET("", Index)

	subscribe.Register(router.Group("subscribe"))
	ws.Register(router.Group("ws"))

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}

}

func Index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
