package blive

import (
	live "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/go-playground/assert/v2"
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

	cancel, err := LaunchLiveServer(24643640, func(data *LiveInfo, msg live.Msg) {
		t.Log(data, msg)
	})

	if err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second * 15)
	cancel()
	<-time.After(time.Second * 3)
}

func init() {
	_ = database.StartDB()
}
