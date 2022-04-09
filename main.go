package main

import (
	"flag"
	"fmt"
	"github.com/eric2788/biligo-live-ws/controller/subscribe"
	ws "github.com/eric2788/biligo-live-ws/controller/websocket"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/eric2788/biligo-live-ws/services/updater"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
)

var release = flag.Bool("release", os.Getenv("GIN_MODE") == "release", "set release mode")
var port = flag.Int("port", 8080, "set the websocket port")

func main() {

	flag.Parse()

	log.Infof("biligo-live-ws version %v", updater.VersionTag)

	if *release {
		gin.SetMode(gin.ReleaseMode)
		blive.Debug = false
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("正在初始化數據庫...")
	if err := database.StartDB(); err != nil {
		log.Fatalf("初始化數據庫時出現嚴重錯誤: %v", err)
	} else {
		log.Info("數據庫已成功初始化。")
	}

	router := gin.Default()

	router.Use(CORS())
	router.Use(ErrorHandler)

	router.GET("", Index)
	router.POST("validate", ValidateProcess)

	subscribe.Register(router.Group("subscribe"))
	ws.Register(router.Group("ws"))

	port := fmt.Sprintf(":%d", *port)

	log.Infof("使用端口 %s\n", port)

	go updater.StartUpdater()

	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}

}

func Index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
