package main

import (
	"flag"
	"fmt"
	"github.com/eric2788/biligo-live-ws/controller/subscribe"
	ws "github.com/eric2788/biligo-live-ws/controller/websocket"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

var release = flag.Bool("release", os.Getenv("GIN_MODE") == "release", "set release mode")
var port = flag.Int("port", 8080, "set the websocket port")

func main() {

	flag.Parse()

	log.Printf("biligo-live-ws version %v", "0.1.8")

	if *release {
		gin.SetMode(gin.ReleaseMode)
		blive.Debug = false
	}

	router := gin.Default()

	router.Use(CORS())
	router.Use(ErrorHandler)

	router.GET("", Index)
	router.POST("validate", ValidateProcess)

	subscribe.Register(router.Group("subscribe"))
	ws.Register(router.Group("ws"))

	port := fmt.Sprintf(":%d", *port)

	log.Printf("使用端口 %s\n", port)

	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}

}

func Index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
