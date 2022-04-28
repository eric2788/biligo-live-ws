package blive

import (
	"context"
	"errors"
	set "github.com/deckarep/golang-set"
	biligo "github.com/eric2788/biligo-live"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"time"
)

var (
	listening          = set.NewSet()
	shortRoomListening = set.NewSet()
	excepted           = set.NewSet()
	liveFetch          = set.NewSet()
	coolingDown        = set.NewSet()

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

func LaunchLiveServer(wg *sync.WaitGroup, room int64, handle func(data *LiveInfo, msg biligo.Msg), cancelGet func(context.CancelFunc)) {

	defer wg.Done()

	log.Debugf("[%v] 正在獲取直播資訊...", room)

	liveInfo, err := GetLiveInfo(room) // 獲取直播資訊

	log.Debugf("[%v] 獲取直播資訊成功。", room)

	if err != nil {

		if err == ErrTooFast {
			// 假設為已添加監聽以防止重複監聽
			coolingDown.Add(room)
			go func() {
				cool := time.Minute*10 + time.Second*time.Duration(len(coolingDown.ToSlice()))
				log.Warnf("將於 %v 後再嘗試監聽直播: %d", shortDur(cool), room)
				// 十分鐘冷卻後再重試
				<-time.After(cool)
				coolingDown.Remove(room)
			}()
		}

		log.Errorf("[%v] 獲取直播資訊失敗: %v", room, err)
		return
	}

	realRoom := liveInfo.RoomId

	// 監聽房間為短號
	if room != realRoom {

		// 添加到映射
		ShortRoomMap.Store(realRoom, room)

		// 真正房間號已經在監聽
		if listening.Contains(realRoom) {
			log.Infof("檢測到 %v 為短號，真正房間號為 %v 且正在監聽中。", room, realRoom)
			shortRoomListening.Add(room)
			return
		}

	}

	removeListen := func() {
		listening.Remove(realRoom)
		if room != realRoom {
			shortRoomListening.Remove(room)
		}
		if shortRoom, ok := ShortRoomMap.Load(realRoom); ok {
			shortRoomListening.Remove(shortRoom)
		}
	}

	live := biligo.NewLive(false, 30*time.Second, 0, func(err error) {
		log.Fatal(err)
	})

	log.Debugf("[%v] 正在連接到彈幕伺服器...", room)

	if err := live.Conn(websocket.DefaultDialer, biligo.WsDefaultHost); err != nil {
		log.Warn("連接伺服器時出現錯誤: ", err)
		return
	}

	log.Debugf("[%v] 連接到彈幕伺服器成功。", room)

	ctx, stop := context.WithCancel(context.Background())

	go func() {

		if err := live.Enter(ctx, realRoom, "", 0); err != nil {
			log.Warnf("監聽房間 %v 時出現錯誤: %v\n", realRoom, err)
			removeListen()
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
				removeListen()
				return
			}
		}
	}()

	listening.Add(realRoom)
	if room != realRoom {
		log.Infof("%v 為短號，已新增真正的房間號 %v => %v 作為監聽。", room, room, realRoom)
		shortRoomListening.Add(room)
	}

	cancelGet(stop)
}

func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
