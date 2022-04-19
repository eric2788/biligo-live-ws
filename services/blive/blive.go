package blive

import (
	"context"
	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/sirupsen/logrus"
	"time"
)

var log = logrus.WithField("service", "blive")
var stopMap = make(map[int64]context.CancelFunc)

func SubscribedRoomTracker(handleWs func(int64, *LiveInfo, live.Msg)) {
	log.Info("已啟動房間訂閱監聽。")
	for {
		time.Sleep(time.Second * 5)

		rooms := subscriber.GetAllRooms()

		for toListen := range rooms.Difference(listening).Iter() {
			if excepted.Contains(toListen) {
				continue
			}
			room := toListen.(int64)

			log.Info("正在啟動監聽房間: ", room)

			stop, err := LaunchLiveServer(room, func(data *LiveInfo, msg live.Msg) {
				handleWs(room, data, msg)
			})

			if err == nil {
				stopMap[room] = stop
			}
		}

		for toStop := range listening.Difference(rooms).Iter() {
			room := toStop.(int64)

			if stop, ok := stopMap[room]; ok {

				log.Info("正在中止監聽房間: ", room)

				stop()
				delete(stopMap, room)
			}
		}
	}
}
