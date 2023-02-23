package websocket

import (
	"github.com/gorilla/websocket"
)

type WriteBuffer struct {
	conn   *websocket.Conn
	buffer []byte
}

var (
	channelMap = make(map[string]chan *WriteBuffer)
)

func insertBuffer(identifier string, conn *websocket.Conn, buffer []byte) {
	channelMap[identifier] <- &WriteBuffer{
		conn:   conn,
		buffer: buffer,
	}
}

func startWriter(identifier string) {
	if _, ok := channelMap[identifier]; ok {
		// delete old
		close(channelMap[identifier])
		delete(channelMap, identifier)
	}
	channel := make(chan *WriteBuffer, 2048)
	channelMap[identifier] = channel
	for {
		select {
		case buffer, ok := <-channel:
			if !ok {
				return
			}
			if err := buffer.conn.WriteMessage(websocket.TextMessage, buffer.buffer); err != nil {
				log.Warnf("向 用戶 %v 發送直播數據時出現錯誤: (%T)%v\n", identifier, err, err)
				log.Warnf("關閉對用戶 %v 的連線。", identifier)
				_ = buffer.conn.Close()
				// 客戶端非正常關閉連接
				HandleClose(identifier)
				return
			}
		}
	}
}
