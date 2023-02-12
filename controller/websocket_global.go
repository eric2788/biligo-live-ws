package controller

import (
	"encoding/json"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func openGlobalWebsocket(c *gin.Context) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if os.Getenv("RESTRICT_GLOBAL") != "" {
				return c.Query("token") == os.Getenv("RESTRICT_GLOBAL")
			}
			return true
		},
		ReadBufferSize:  64,
		WriteBufferSize: 2048,
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}

	identifier := fmt.Sprintf("%v@%v", c.ClientIP(), "global")

	// 客戶端正常關閉連接
	ws.SetCloseHandler(func(code int, text string) error {
		log.Infof("已關閉對 %v 的 Websocket 連接: (%v) %v", identifier, code, text)
		subscriber.RemoveSubscriber(identifier)
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})

	// ignore the subscribed rooms
	sub := subscriber.GetSubscriber(identifier)
	sub.Websocket = ws
	sub.Global = true
}

func writeGlobalMessage(identifier string, sub *subscriber.Subscriber, data BLiveData) error {

	if !sub.IsConnected() {
		return fmt.Errorf("用戶 %v 的連接已關閉", identifier)
	}

	sub.Locker.Lock()
	defer sub.Locker.Unlock()

	con := sub.Websocket
	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if err = con.WriteMessage(websocket.TextMessage, byteData); err != nil {
		log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
		log.Warnf("關閉對用戶 %v 的連線。", identifier)
		_ = con.Close()
		subscriber.RemoveSubscriber(identifier)
	}

	return nil
}
