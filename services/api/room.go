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

	if res, ok := roomCaches.Load(room); ok {
		return res.(*RoomInfo), nil
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

func RoomExist(room int64) (bool, error) {
	res, err := GetRoomInfo(room)

	if err != nil {
		return false, err
	}
	return res.Code == 0, nil
}
