package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/blive"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sirupsen/logrus"
)

var (
	websocketTable = cmap.New[*websocket.Conn]()
	log            = logrus.WithField("controller", "websocket")
)

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
		ReadBufferSize:  64,
		WriteBufferSize: 2048,
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

	websocketTable.Set(identifier, ws)

	// 先前尚未有訂閱
	if _, subBefore := subscriber.Get(identifier); !subBefore {
		// 使用空值防止啟動訂閱過期
		subscriber.Update(identifier, []int64{})
	}

	subscriber.AddConnected(identifier)
	go startWriter(identifier)
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
	globalWebSockets.IterCb(func(id string, conn *websocket.Conn) {
		if err := writeGlobalMessage(id, conn, bLiveData); err != nil {
			log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", id, err, err)
		}
	})

}

func writeMessage(identifier string, data BLiveData) error {
	con, ok := websocketTable.Get(identifier)

	if !ok {
		//log.Infof("用戶 %v 尚未連接到WS，略過發送。\n", identifier)
		return nil
	}

	byteData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	go insertBuffer(identifier, con, byteData)
	return nil
}

func HandleClose(identifier string) {
	websocketTable.Remove(identifier)
	subscriber.RemoveConnected(identifier)
	// 等待五分鐘，如果五分鐘後沒有重連則刪除訂閱記憶
	subscriber.ExpireAfter(identifier, time.NewTimer(time.Minute*5))
}

type BLiveData struct {
	Command  string          `json:"command"`
	LiveInfo *blive.LiveInfo `json:"live_info"`
	Content  interface{}     `json:"content"`
}
