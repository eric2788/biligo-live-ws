package api

import (
	"encoding/json"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/database"
	"io"
	"log"
	"net/http"
)

const RoomInfoApi string = "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%v"

func GetRoomInfo(room int64) (*RoomInfo, error) {
	return GetRoomInfoWithOption(room, false)
}

func GetRoomInfoWithOption(room int64, forceUpdate bool) (*RoomInfo, error) {

	dbKey := fmt.Sprintf("room:%v", room)

	if !forceUpdate {
		var roomInfo = &RoomInfo{}
		if err := database.GetFromDB(dbKey, roomInfo); err == nil {
			return roomInfo, nil
		} else {
			if e, ok := err.(*database.EmptyError); ok {
				log.Printf("%v, 使用 web api 更新", e)
			} else {
				log.Printf("從數據庫獲取房間資訊 %v 時出現錯誤: %v, 使用 web api 更新", room, err)
			}
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

	if err := database.PutToDB(dbKey, roomInfo); err != nil {
		log.Printf("從數據庫更新房間資訊 %v 時出現錯誤: %v", room, err)
	} else {
		log.Printf("成功更新房間資訊 %v 到數據庫", room)
	}
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
