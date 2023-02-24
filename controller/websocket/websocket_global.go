package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var globalWebSockets = cmap.New[*websocket.Conn]()

func OpenGlobalWebSocket(c *gin.Context) {

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
		globalWebSockets.Remove(identifier)
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})

	globalWebSockets.Set(identifier, ws)

	go startWriter(identifier)

}

func writeGlobalMessage(identifier string, con *websocket.Conn, data BLiveData) error {

	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	go insertBuffer(identifier, con, byteData)
	return nil
}
