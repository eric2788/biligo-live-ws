package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const RoomInfoApi string = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%v"

var roomCaches = sync.Map{}

func GetRoomInfo(room int64) (*RoomInfo, error) {
	return GetRoomInfoWithOption(room, false)
}

func GetRoomInfoWithOption(room int64, forceUpdate bool) (*RoomInfo, error) {

	if !forceUpdate {
		if res, ok := roomCaches.Load(room); ok {
			return res.(*RoomInfo), nil
		}
	}

	resp, err := http.Get(fmt.Sprintf(RoomInfoApi, room))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var v1resp V1Resp

	if err := json.Unmarshal(body, &v1resp); err != nil {
		return nil, err
	}

	if v1resp.Code != 0 {
		return &RoomInfo{V1Resp: v1resp}, nil
	}

	var roomInfo RoomInfo
	if err := json.Unmarshal(body, &roomInfo); err != nil {
		return nil, err
	}

	roomCaches.Store(room, &roomInfo)
	return &roomInfo, nil

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
