package api

import (
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestGetRoomInfo(t *testing.T) {
	if roomInfo, err := GetRoomInfo(545); err == nil {
		assert.Equal(t, roomInfo.Code, 0)
		assert.Equal(t, roomInfo.Data.ShortId, 545)
		assert.Equal(t, roomInfo.Data.RoomId, int64(573893))
		assert.Equal(t, roomInfo.Data.Uid, int64(15641218))
	} else {
		t.Fatal(err)
	}
}

func TestGetUserInfo(t *testing.T) {
	if userInfo, err := GetUserInfo(1838190318, false); err == nil {
		assert.Equal(t, userInfo.Code, 0)
		assert.Equal(t, userInfo.Data.Name, "魔狼咪莉娅")
		assert.Equal(t, userInfo.Data.Mid, int64(1838190318))
		assert.Equal(t, userInfo.Data.Sex, "保密")
	} else {
		t.Fatal(err)
	}
}

func init() {
	_ = database.StartDB()
}
