package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
)

var globalWebSockets = sync.Map{}

func OpenGlobalWebSocket(c *gin.Context) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if os.Getenv("RESTRICT_GLOBAL") != "" {
				return c.Query("token") == os.Getenv("RESTRICT_GLOBAL")
			}
			return true
		},
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
		globalWebSockets.Delete(identifier)
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})

	globalWebSockets.Store(identifier, &WebSocket{ws: ws})

	go func() {
		for {
			// 接收客戶端關閉訊息
			if _, _, err = ws.NextReader(); err != nil {
				if err := ws.Close(); err != nil {
					log.Warnf("關閉用戶 %v 的 WebSocket 時發生錯誤: %v", identifier, err)
				}
				return
			}
		}
	}()
}

func writeGlobalMessage(identifier string, socket *WebSocket, data BLiveData) error {

	defer socket.mu.Unlock()
	socket.mu.Lock()

	con := socket.ws
	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if err = con.WriteMessage(websocket.TextMessage, byteData); err != nil {
		log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
		log.Warnf("關閉對用戶 %v 的連線。", identifier)
		_ = con.Close()
		globalWebSockets.Delete(identifier)
	}

	return nil
}
