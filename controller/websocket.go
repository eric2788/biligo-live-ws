package controller

import (
	"encoding/json"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"net/http"
	"time"

	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func WebSocket(gp *gin.RouterGroup) {
	gp.GET("", openWebsocket)
	gp.GET("/global", openGlobalWebsocket)
	go blive.SubscribedRoomTracker(handleBLiveMessage)
}

func openWebsocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
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

	// get identifier
	// special case for websocket, not like http
	// not going to use c.getString(string)

	id, ok := c.GetQuery("id")

	// 沒有 id 則為 anonymous
	if !ok {
		id = "anonymous"
	}
	identifier := fmt.Sprintf("%v@%v", c.ClientIP(), id)

	// ===========================

	sub := subscriber.GetSubscriber(identifier)
	sub.Websocket = ws

	// 客戶端正常關閉連接
	ws.SetCloseHandler(func(code int, text string) error {
		logger.Infof("已關閉對 %v 的 Websocket 連接: (%v) %v", identifier, code, text)
		HandleClose(identifier, sub)
		return ws.WriteMessage(websocket.CloseMessage, nil)
	})
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
		logger.Warnf("序列化 原始數據內容 時出現錯誤: %v, 將轉換為 string", err)
		content = string(raw)
	}

	bLiveData := BLiveData{
		Command:  msg.Cmd(),
		LiveInfo: info,
		Content:  content,
	}

	// 訂閱用戶
	for identifier, sub := range subscriber.GetAllSubscribers(room) {
		if err := writeMessage(identifier, sub, bLiveData); err != nil {
			logger.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
		}
	}

	// 短號用戶
	if shortRoomId, ok := blive.ShortRoomMap.Load(room); ok {
		for identifier, sub := range subscriber.GetAllSubscribers(shortRoomId.(int64)) {
			if err := writeMessage(identifier, sub, bLiveData); err != nil {
				logger.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
			}
		}
	}

	// 全局用戶
	for id, sub := range subscriber.GetAllGlobalSubscribers() {
		if err := writeGlobalMessage(id, sub, bLiveData); err != nil {
			logger.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", id, err, err)
		}
	}

}

func writeMessage(identifier string, sub *subscriber.Subscriber, data BLiveData) error {

	if !sub.IsConnected() {
		return fmt.Errorf("用戶 %v 的連接已關閉", identifier)
	}

	defer sub.Locker.Unlock()
	sub.Locker.Lock()

	con := sub.Websocket

	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if err = con.WriteMessage(websocket.TextMessage, byteData); err != nil {
		logger.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
		logger.Warnf("關閉對用戶 %v 的連線。", identifier)
		_ = con.Close()
		// 客戶端非正常關閉連接
		HandleClose(identifier, sub)
	}

	return nil
}

func HandleClose(identifier string, sub *subscriber.Subscriber) {
	sub.Websocket = nil
	// 等待五分鐘，如果五分鐘後沒有重連則刪除訂閱記憶
	// 由於斷線的時候已經有訂閱列表，因此此方法不會檢查是否有訂閱列表
	go subscriber.ActiveExpire(identifier, time.Minute*5)
}

type BLiveData struct {
	Command  string          `json:"command"`
	LiveInfo *blive.LiveInfo `json:"live_info"`
	Content  interface{}     `json:"content"`
}
