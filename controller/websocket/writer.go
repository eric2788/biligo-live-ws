package websocket

import (
	"context"
	"strings"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type (
	WsChannel struct {
		writer chan *WriteBuffer
		ctx    context.Context
	}

	WriteBuffer struct {
		conn   *websocket.Conn
		buffer []byte
	}
)

var (
	channelMap = cmap.New[*WsChannel]()
)

func insertBuffer(identifier string, conn *websocket.Conn, buffer []byte) {

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic when sending data to %s: %v", identifier, err)
		}
	}()

	channel, ok := channelMap.Get(identifier)

	if !ok {
		return
	}

	channel.writer <- &WriteBuffer{
		conn:   conn,
		buffer: buffer,
	}
}

func startWriter(identifier string) {

	if channel, ok := channelMap.Get(identifier); ok {

		// delete first, then close channel
		channelMap.Remove(identifier)
		close(channel.writer)

		// 等待上一个 writer 全部写入完毕
		<-channel.ctx.Done()
		log.Infof("成功關閉用戶 %v 的寫入器", identifier)
	}

	var buffer int

	if strings.HasSuffix(identifier, "global") {
		buffer = 150000
	} else {
		buffer = 50000
	}

	ctx, cancel := context.WithCancel(context.Background())

	channel := make(chan *WriteBuffer, buffer)

	log.Infof("為用戶 %v 啟動寫入器, 緩衝大小為 %db", identifier, buffer)

	channelMap.Set(identifier, &WsChannel{
		writer: channel,
		ctx:    ctx,
	})

	defer cancel()

	for buffer := range channel {
		if err := buffer.conn.WriteMessage(websocket.TextMessage, buffer.buffer); err != nil {
			log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
			log.Warnf("關閉對用戶 %v 的連線。", identifier)
			_ = buffer.conn.Close()
			// 客戶端非正常關閉連接
			HandleClose(identifier)
			channelMap.Remove(identifier)
			return
		}
	}
}
