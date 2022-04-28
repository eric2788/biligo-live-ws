package blive

import (
	"context"
	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var log = logrus.WithField("service", "blive")
var stopMap = sync.Map{}

func SubscribedRoomTracker(handleWs func(int64, *LiveInfo, live.Msg)) {
	log.Info("已啟動房間訂閱監聽。")
	wg := &sync.WaitGroup{}
	for {
		time.Sleep(time.Second * 5)

		rooms := subscriber.GetAllRooms()

		log.Debug("房間訂閱: ", rooms.ToSlice())
		log.Debug("正在監聽: ", listening.ToSlice())

		for toListen := range rooms.Difference(listening).Iter() {

			if excepted.Contains(toListen) {
				log.Debugf("房間 %v 已排除", toListen)
				continue
			}
			// 已經啟動監聽的短號
			if shortRoomListening.Contains(toListen) {
				log.Debugf("房間 %v 已經啟動短號監聽", toListen)
				continue
			}

			// 冷卻時暫不監聽直播
			if coolingDown.Contains(toListen) {
				log.Debugf("房間 %v 在冷卻時暫不監聽直播", toListen)
				continue
			}

			room := toListen.(int64)

			log.Info("正在啟動監聽房間: ", room)

			wg.Add(1)
			go LaunchLiveServer(wg, room,
				func(data *LiveInfo, msg live.Msg) {
					handleWs(room, data, msg)
				}, func(stop context.CancelFunc, err error) {
					if err == nil && stop != nil {
						stopMap.Store(room, stop)
					} else {
						listening.Remove(room)
						if short, ok := ShortRoomMap.Load(room); ok {
							shortRoomListening.Remove(short)
						}
						log.Warnf("已移除房間 %v 的監聽狀態", room)
					}
				})
			listening.Add(room)
		}

		wg.Wait()

		for short := range shortRoomListening.Iter() {
			rooms.Add(short)
		}

		for toStop := range listening.Difference(rooms).Iter() {
			room := toStop.(int64)

			if stop, ok := stopMap.LoadAndDelete(room); ok {
				log.Info("正在中止監聽房間: ", room)
				stop.(context.CancelFunc)()
			}
		}

	}
}
