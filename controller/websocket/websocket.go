package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	live "github.com/iyear/biligo-live"
	"log"
	"net/http"
	"sync"
	"time"
)

var websocketTable = sync.Map{}

type WebSocket struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func Register(gp *gin.RouterGroup) {
	gp.GET("", OpenWebSocket)

	go blive.SubscribedRoomTracker(handleBLiveMessage)
}

func OpenWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ws.SetCloseHandler(func(code int, text string) error {
		log.Printf("已關閉對 %v 的 Websocket 連接: (%v) %v", c.ClientIP(), code, text)
		websocketTable.Delete(c.ClientIP())
		// 等待五分鐘，如果五分鐘後沒有重連則刪除訂閱記憶
		subscriber.ExpireAfter(c.ClientIP(), time.After(time.Minute*5))
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})

	websocketTable.Store(c.ClientIP(), &WebSocket{ws: ws})

	// 中止五分鐘後清除訂閱記憶
	subscriber.CancelExpire(c.ClientIP())

	go func() {
		for {
			// 接收客戶端關閉訊息
			if _, _, err = ws.NextReader(); err != nil {
				if err := ws.Close(); err != nil {
					log.Printf("關閉用戶 %v 的 WebSocket 時發生錯誤: %v", c.ClientIP(), err)
				}
				return
			}
		}
	}()
}

func handleBLiveMessage(room int64, info blive.LiveInfo, msg live.Msg) {

	raw := msg.Raw()

	// 人氣值轉換為 json string
	if reply, ok := msg.(*live.MsgHeartbeatReply); ok {
		hot := reply.GetHot()
		raw = []byte(fmt.Sprintf("{\"popularity\": %v}", hot))
	}

	bLiveData := BLiveData{
		Command:  msg.Cmd(),
		LiveInfo: info,
		Content:  raw,
	}

	// if no comment will spam
	//if _, ok := msg.(*live.MsgHeartbeatReply); !ok { // 非 heartbeat 訊息
	//	body := string(msg.Raw())
	//	log.Printf("從 %v 收到訊息: %v\n", room, body)
	//}

	for _, ip := range subscriber.GetAllSubscribers(room) {
		if err := writeMessage(ip, bLiveData); err != nil {
			log.Printf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", ip, err, err)
		}
	}

}

func writeMessage(ip string, data BLiveData) error {
	conn, ok := websocketTable.Load(ip)

	if !ok {
		//log.Printf("用戶 %v 尚未連接到WS，略過發送。\n", ip)
		return nil
	}

	webSocket := conn.(*WebSocket)
	defer webSocket.mu.Unlock()
	webSocket.mu.Lock()

	con := webSocket.ws
	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if err = con.WriteMessage(websocket.TextMessage, byteData); err != nil {
		return err
	}

	return nil
}

type BLiveData struct {
	Command  string         `json:"command"`
	LiveInfo blive.LiveInfo `json:"live_info"`
	Content  []byte         `json:"content"`
}
