package websocket

import (
	"encoding/json"
	"fmt"
	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	websocketTable = sync.Map{}
	log            = logrus.WithField("controller", "websocket")
)

type WebSocket struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func Register(gp *gin.RouterGroup) {
	gp.GET("", OpenWebSocket)
	gp.GET("/global", OpenGlobalWebSocket)
	go blive.SubscribedRoomTracker(handleBLiveMessage)
}

func OpenWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 獲取辨識 Id
	id, ok := c.GetQuery("id")

	// 沒有 id 則為 anonymous
	if !ok {
		id = "anonymous"
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}

	identifier := fmt.Sprintf("%v@%v", c.ClientIP(), id)

	// 客戶端正常關閉連接
	ws.SetCloseHandler(func(code int, text string) error {
		log.Infof("已關閉對 %v 的 Websocket 連接: (%v) %v", identifier, code, text)
		HandleClose(identifier)
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})

	websocketTable.Store(identifier, &WebSocket{ws: ws})

	// 先前尚未有訂閱
	if _, subBefore := subscriber.Get(identifier); !subBefore {
		// 使用空值防止啟動訂閱過期
		subscriber.Update(identifier, []int64{})
	}

	// 中止五分鐘後清除訂閱記憶
	subscriber.CancelExpire(identifier)

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

func handleBLiveMessage(room int64, info *blive.LiveInfo, msg live.Msg) {

	raw := msg.Raw()

	// 人氣值轉換為 json string
	if reply, ok := msg.(*live.MsgHeartbeatReply); ok {
		hot := reply.GetHot()
		raw = []byte(fmt.Sprintf("{\"popularity\": %v}", hot))
	}

	var content interface{}

	if err := json.Unmarshal(raw, &content); err != nil {
		log.Warnf("序列化 原始數據內容 時出現錯誤: %v, 將轉換為 string", err)
		content = string(raw)
	}

	bLiveData := BLiveData{
		Command:  msg.Cmd(),
		LiveInfo: info,
		Content:  content,
	}

	// 訂閱用戶
	for _, identifier := range subscriber.GetAllSubscribers(room) {
		if err := writeMessage(identifier, bLiveData); err != nil {
			log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
		}
	}

	// 短號用戶
	if shortRoomId, ok := blive.ShortRoomMap.Load(room); ok {

		for _, identifier := range subscriber.GetAllSubscribers(shortRoomId.(int64)) {
			if err := writeMessage(identifier, bLiveData); err != nil {
				log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
			}
		}

	}

	// 全局用戶
	for id, conn := range globalWebSockets {
		if err := writeGlobalMessage(id, conn, bLiveData); err != nil {
			log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", id, err, err)
		}
	}

}

func writeMessage(identifier string, data BLiveData) error {
	conn, ok := websocketTable.Load(identifier)

	if !ok {
		//log.Infof("用戶 %v 尚未連接到WS，略過發送。\n", identifier)
		return nil
	}

	socket := conn.(*WebSocket)

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
		// 客戶端非正常關閉連接
		HandleClose(identifier)
	}

	return nil
}

func HandleClose(identifier string) {
	websocketTable.Delete(identifier)
	// 等待五分鐘，如果五分鐘後沒有重連則刪除訂閱記憶
	// 由於斷線的時候已經有訂閱列表，因此此方法不會檢查是否有訂閱列表
	subscriber.ExpireAfterWithCheck(identifier, time.After(time.Minute*5), false)
}

type BLiveData struct {
	Command  string          `json:"command"`
	LiveInfo *blive.LiveInfo `json:"live_info"`
	Content  interface{}     `json:"content"`
}
