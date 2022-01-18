package main

import (
	"fmt"
	"github.com/eric2788/biligo-live-ws/controller/subscribe"
	ws "github.com/eric2788/biligo-live-ws/controller/websocket"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {

	log.Printf("biligo-live-ws version %v", "0.1.6")

	router := gin.Default()

	router.Use(CORS())
	router.Use(ErrorHandler)

	router.GET("", Index)
	router.POST("validate", ValidateProcess)

	subscribe.Register(router.Group("subscribe"))
	ws.Register(router.Group("ws"))

	port := ":8080"

	if len(os.Args) > 1 {
		port = fmt.Sprintf(":%v", os.Args[1])
	}

	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}

}

func Index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
