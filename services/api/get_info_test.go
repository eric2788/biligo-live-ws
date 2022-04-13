package api

import (
	"encoding/json"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/go-playground/assert/v2"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strings"
	"testing"
)

func TestGetRoomInfo(t *testing.T) {
	if roomInfo, err := GetRoomInfo(545); err == nil {
		assert.Equal(t, roomInfo.Code, 0)
		assert.Equal(t, roomInfo.Data.ShortId, 545)
		assert.Equal(t, roomInfo.Data.RoomId, int64(573893))
		assert.Equal(t, roomInfo.Data.Uid, int64(15641218))
		assert.MatchRegex(t, roomInfo.Data.UserCover, "^https://.*")
	} else {
		t.Fatal(err)
	}
}

func UpdateDbFromHttpToHttps(t *testing.T) {
	database.UpdateDB(func(db *leveldb.DB) error {
		iter := db.NewIterator(util.BytesPrefix([]byte("user:")), nil)
		for iter.Next() {
			var userInfo UserInfo
			err := json.Unmarshal(iter.Value(), &userInfo)
			if err != nil {
				t.Error("error when unmarshall userinfo: ", err)
				continue
			}
			userInfo.Data.Face = strings.Replace(userInfo.Data.Face, "http://", "https://", -1)
			b, err := json.Marshal(userInfo)
			if err != nil {
				t.Errorf("error when marshall userinfo %v: %v", userInfo.Data.Name, err)
				continue
			}
			t.Logf("updating %v", string(iter.Key()))
			err = db.Put(iter.Key(), b, nil)
			if err != nil {
				t.Errorf("error when update user info %v: %v", userInfo.Data.Name, err)
			} else {
				t.Logf("update user info %v success: %v", userInfo.Data.Name, err)
			}
		}
		iter.Release()
		return iter.Error()
	})
}

func TestGetUserInfo(t *testing.T) {
	if userInfo, err := GetUserInfo(1838190318, false); err == nil {
		assert.Equal(t, userInfo.Code, 0)
		assert.Equal(t, userInfo.Data.Name, "魔狼咪莉娅")
		assert.Equal(t, userInfo.Data.Mid, int64(1838190318))
		assert.Equal(t, userInfo.Data.Sex, "保密")
		// assert the url link is start with https
		assert.MatchRegex(t, userInfo.Data.Face, "^https://.*")
	} else {
		t.Fatal(err)
	}
}

func init() {
	_ = database.StartDB()
}
