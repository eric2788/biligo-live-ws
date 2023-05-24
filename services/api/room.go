package api

import (
	"fmt"
	"github.com/eric2788/common-services/bilibili"
	"strings"

	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("service", "api")

func GetRoomInfo(room int64) (*bilibili.RoomInfo, error) {
	return GetRoomInfoWithOption(room, false)
}

func GetRoomInfoCache(room int64) (*bilibili.RoomInfo, error) {

	dbKey := fmt.Sprintf("room:%v", room)

	var roomInfo = &bilibili.RoomInfo{}
	if err := database.GetFromDB(dbKey, roomInfo); err == nil {
		return roomInfo, nil
	} else {
		if _, ok := err.(*database.EmptyError); ok {
			return nil, ErrCacheNotFound
		} else {
			return nil, err
		}
	}

}

func GetRoomInfoWithOption(room int64, forceUpdate bool) (*bilibili.RoomInfo, error) {

	dbKey := fmt.Sprintf("room:%v", room)

	if !forceUpdate {
		if roomInfo, err := GetRoomInfoCache(room); err == nil {
			return roomInfo, nil
		} else {
			if err == ErrCacheNotFound {
				log.Debugf("%v, 正在請求B站 API", err)
			} else {
				log.Warnf("從數據庫獲取房間資訊 %v 時出現錯誤: %v, 正在請求B站 API", room, err)
			}
		}
	}

	roomInfo, err := bilibili.GetRoomInfo(room)
	if err != nil {
		return nil, err
	}

	roomInfo.Data.UserCover = strings.Replace(roomInfo.Data.UserCover, "http://", "https://", -1)

	if err := database.PutToDB(dbKey, roomInfo); err != nil {
		log.Warnf("從數據庫更新房間資訊 %v 時出現錯誤: %v", room, err)
	} else {
		log.Debugf("房間資訊 %v 更新到數據庫成功", room)
	}
	return roomInfo, nil

}

func GetRealRoom(room int64) (int64, error) {
	res, err := GetRoomInfo(room)

	// 错误
	if err != nil {
		return -1, err
	}

	// 房间不存在
	if res.Data == nil {
		return -1, nil
	}

	return res.Data.RoomId, nil // 返回真实房间号

}
