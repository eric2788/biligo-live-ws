package blive

import (
	"context"
	"errors"
	set "github.com/deckarep/golang-set"
	biligo "github.com/eric2788/biligo-live"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var (
	listening = set.NewSet()
	excepted  = set.NewSet()
	liveFetch = set.NewSet()

	ShortRoomMap = sync.Map{}
)

var (
	ErrNotFound = errors.New("房間不存在")
	ErrTooFast  = errors.New("請求頻繁")
)

func GetListening() []interface{} {
	return listening.ToSlice()
}

func coolDownLiveFetch(room int64) {
	liveFetch.Add(room)
	<-time.After(time.Minute * 5)
	liveFetch.Remove(room)
}

var Debug = true

func LaunchLiveServer(room int64, handle func(data *LiveInfo, msg biligo.Msg)) (context.CancelFunc, error) {

	liveInfo, err := GetLiveInfo(room) // 獲取直播資訊

	if err != nil {

		if err == ErrTooFast {
			// 假設為已添加監聽以防止重複監聽
			listening.Add(room)
			go func() {
				log.Warnf("將於十分鐘後再嘗試監聽直播: %d", room)
				// 十分鐘冷卻後再重試
				<-time.After(time.Minute * 10)
				listening.Remove(room)
			}()
		}

		return nil, err
	}

	realRoom := liveInfo.RoomId

	// 監聽房間為短號
	if room != realRoom {

		// 添加到映射
		ShortRoomMap.Store(realRoom, room)

		// 真正房間號已經在監聽
		if listening.Contains(realRoom) {
			log.Infof("檢測到 %v 為短號，真正房間號為 %v 且正在監聽中。", room, realRoom)
			excepted.Add(room)
			return nil, errors.New("此房間已經在監聽")
		}

	}

	live := biligo.NewLive(Debug, 30*time.Second, 0, func(err error) {
		log.Fatal(err)
	})

	if err := live.Conn(websocket.DefaultDialer, biligo.WsDefaultHost); err != nil {
		log.Warn("連接伺服器時出現錯誤: ", err)
		return nil, err
	}

	ctx, stop := context.WithCancel(context.Background())

	go func() {

		if err := live.Enter(ctx, realRoom, "", 0); err != nil {
			log.Warnf("監聽房間 %v 時出現錯誤: %v\n", realRoom, err)
			listening.Remove(realRoom)
			stop()
		}

	}()

	go func() {
		for {
			select {
			case tp := <-live.Rev:
				if tp.Error != nil {
					log.Info(tp.Error)
					continue
				}
				// 開播 !?
				if _, ok := tp.Msg.(*biligo.MsgLive); ok {

					// 更新直播資訊只做一次
					if !liveFetch.Contains(realRoom) {
						go coolDownLiveFetch(realRoom)
						log.Infof("房間 %v 開播，正在更新直播資訊...\n", realRoom)
						// 更新一次直播资讯
						UpdateLiveInfo(liveInfo, realRoom)
					}

					// 但開播指令推送多次保留
				}
				handle(liveInfo, tp.Msg)
			case <-ctx.Done():
				log.Infof("房間 %v 監聽中止。\n", realRoom)
				listening.Remove(realRoom)
				return
			}
		}
	}()

	listening.Add(room)
	if room != realRoom {
		log.Infof("%v 為短號，已新增真正的房間號 %v => %v 作為監聽。", room, room, realRoom)
		listening.Add(realRoom)
	}
	return stop, nil
}
