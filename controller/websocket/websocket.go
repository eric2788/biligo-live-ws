package websocket

import (
	"encoding/json"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	live "github.com/iyear/biligo-live"
	"log"
	"sync"
	"time"
)

var websocketTable = sync.Map{}

func Register(gp *gin.RouterGroup) {
	gp.GET("", OpenWebSocket)
	go blive.SubscribedRoomTracker(handleBLiveMessage)
}

func OpenWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}
	websocketTable.Store(c.ClientIP(), ws)
	// 中止五分鐘後清除訂閱記憶
	subscriber.CancelExpire(c.ClientIP())
}

func handleBLiveMessage(room int64, info blive.LiveInfo, msg live.Msg) {

	bLiveData := BLiveData{
		Command:  msg.Cmd(),
		LiveInfo: info,
		Content:  msg.Raw(),
	}

	// if no comment will spam
	if _, ok := msg.(*live.MsgHeartbeatReply); !ok { // 非 heartbeat 訊息
		body := string(msg.Raw())
		log.Printf("從 %v 收到訊息: %v\n", room, body)
	}

	for _, ip := range subscriber.GetAllSubscribers(room) {
		if err := writeMessage(ip, bLiveData); err != nil {
			log.Printf("向 用戶 %v 發送直播數據時出現錯誤: %v\n", ip, err)
		}
	}

}

func writeMessage(ip string, data BLiveData) error {
	conn, ok := websocketTable.Load(ip)

	if !ok {
		log.Printf("用戶 %v 尚未連接到WS，略過發送。\n", ip)
		return nil
	}

	con := conn.(*websocket.Conn)

	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if err = con.WriteMessage(websocket.TextMessage, byteData); err != nil {
		switch err.(type) {
		case *websocket.CloseError: // websocket 關閉
			log.Printf("用戶 %v 已斷開WS連接。\n", ip)
			websocketTable.Delete(ip)
			// 等待五分鐘，如果五分鐘後沒有重連則刪除訂閱記憶
			subscriber.ExpireAfter(ip, time.After(time.Minute*5))
		default:
			return err
		}
	}

	return nil
}

type BLiveData struct {
	Command  string         `json:"command"`
	LiveInfo blive.LiveInfo `json:"live_info"`
	Content  []byte         `json:"content"`
}
