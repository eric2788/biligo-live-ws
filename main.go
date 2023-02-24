package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/eric2788/biligo-live-ws/controller/listening"
	"github.com/eric2788/biligo-live-ws/controller/subscribe"
	ws "github.com/eric2788/biligo-live-ws/controller/websocket"
	"github.com/eric2788/biligo-live-ws/middleware"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/eric2788/biligo-live-ws/services/updater"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var release = flag.Bool("release", os.Getenv("GIN_MODE") == "release", "set release mode")
var port = flag.Int("port", 8080, "set the websocket port")

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

	ts := time.Now().Format("2006-01-02 15:04:05")
	// create log with unix timestamp
	logFile, err := os.OpenFile("./cache/networkLog/"+ts+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0775)
	if err != nil {
		log.Fatalf("创建 /cache/networkLog/%s.log 文件时错误: %v", ts, err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	gin.DefaultWriter = mw

	router := gin.New()

	router.Use(func(c *gin.Context) {
		if os.Getenv("NO_LISTENING_LOG") == "true" && strings.HasPrefix(c.Request.URL.Path, "/listening") {
			c.Next()
			return
		}
		gin.Logger()(c)
	})

	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.Identifier())

	router.GET("", Index)
	router.POST("validate", ValidateProcess)

	subscribe.Register(router.Group("subscribe"))
	ws.Register(router.Group("ws"))
	listening.Register(router.Group("listening"))
	pprof.Register(router)

	port := fmt.Sprintf(":%d", *port)

	log.Infof("使用端口 %s\n", port)

	go updater.StartUpdater()

	if err := router.Run(port); err != nil {
		log.Fatal(err)
	}

	if err := database.CloseDB(); err != nil {
		log.Errorf("關閉數據庫時錯誤: %v", err)
	}
}

func Index(c *gin.Context) {
	c.IndentedJSON(200, gin.H{
		"status": "working",
	})
}
