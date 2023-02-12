package blive

import (
	"context"
	"sync"
	"testing"
	"time"

	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/eric2788/biligo-live-ws/services/subscriber"
	"github.com/go-playground/assert/v2"
	"github.com/sirupsen/logrus"
)

func TestGetLiveInfo(t *testing.T) {

	info, err := GetLiveInfo(24643640)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, info.UID, int64(1838190318))
	assert.Equal(t, info.Name, "魔狼咪莉娅")
}

func TestSubscribedRoomTracker(t *testing.T) {
	subscriber.GetSubscriber("tester-1").AddSubscribes([]int64{255, 525, 545, 5424})
	subscriber.GetSubscriber("tester-2").AddSubscribes([]int64{573893, 394681, 48743})

	go SubscribedRoomTracker(func(i int64, info *LiveInfo, msg live.Msg) {
		t.Log(i, msg.Cmd())
	})

	<-time.After(time.Second * 15)
}

func TestLaunchLiveServer(t *testing.T) {
	var cancel context.CancelFunc
	wg := sync.WaitGroup{}
	wg.Add(1)
	go LaunchLiveServer(&wg, 24643640, func(data *LiveInfo, msg live.Msg) {
		t.Log(msg.Cmd())
	}, func(stop context.CancelFunc, err error) {
		if err == nil {
			cancel = stop
		} else {
			t.Error(err)
			return
		}
	})

	<-time.After(time.Second * 15)
	cancel()
	<-time.After(time.Second * 3)
}

func TestTimer(t *testing.T) {
	timer := time.NewTimer(time.Second * 5)
	defer timer.Stop()
	<-timer.C
	t.Log("done")
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	_ = database.StartDB()
}
