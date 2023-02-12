package main

import (
	"flag"
	"fmt"
	"github.com/eric2788/biligo-live-ws/controller"
	"github.com/eric2788/biligo-live-ws/middleware"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/eric2788/biligo-live-ws/services/updater"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

var release = flag.Bool("release", os.Getenv("GIN_MODE") == "release", "set release mode")
var port = flag.Int("port", 8080, "set the websocket port")

func index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
func validateSubs(c *gin.Context) {
	subs, ok := subscriber.GetSubscribes(c.GetString("identifier"))
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "尚未訂閱任何的直播房間號"})
		return
	}
	if len(subs) == 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "訂閱列表為空"})
		return
	}

	c.Status(200)
}

func main() {

	flag.Parse()

	log.Infof("biligo-live-ws v%v", updater.VersionTag)

	if *release {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
		log.Debug("啟動debug模式")
	}

	log.Info("正在初始化數據庫...")
	if err := database.StartDB(); err != nil {
		log.Fatalf("初始化數據庫時出現嚴重錯誤: %v", err)
	} else {
		log.Info("數據庫已成功初始化。")
	}

	router := gin.New()

	if os.Getenv("NO_LISTENING_LOG") == "true" {
		router.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/listening") {
				c.Next()
				return
			}
			gin.Logger()(c)
		})
	}

	if os.Getenv("RESET_LOW_LATENCY") == "true" {
		go api.ResetAllLowLatency()
	}

	router.Use(middleware.CORS())
	router.Use(middleware.Identifier())
	router.Use(middleware.ErrorHandler())

	router.GET("", index)
	router.POST("validate", validateSubs)

	controller.Subscribe(router.Group("subscribe"))
	controller.WebSocket(router.Group("ws"))
	controller.Listening(router.Group("listening"))

	port := fmt.Sprintf(":%d", *port)

	log.Infof("使用端口 %s\n", port)

	go debugServe()
	go updater.StartUpdater()

	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}

	if err := database.CloseDB(); err != nil {
		log.Errorf("關閉數據庫時錯誤: %v", err)
	}
}
