package blive

import (
	"context"
	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/go-playground/assert/v2"
	"github.com/sirupsen/logrus"
	"sync"
	"testing"
	"time"
)

func TestGetLiveInfo(t *testing.T) {

	info, err := GetLiveInfo(24643640)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, info.UID, int64(1838190318))
	assert.Equal(t, info.Name, "魔狼咪莉娅")
}

func TestLaunchLiveServer(t *testing.T) {

	var cancel context.CancelFunc
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go LaunchLiveServer(wg, 24643640, func(data *LiveInfo, msg live.Msg) {
		t.Log(data, msg)
	}, func(cancelFunc context.CancelFunc) {
		cancel = cancelFunc
	})

	wg.Wait()

	if cancel == nil {
		t.Fatal("cancel is nil")
	}
	<-time.After(time.Second * 15)
	cancel()
	<-time.After(time.Second * 3)
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	_ = database.StartDB()
}
