package blive

import (
	"context"
	"errors"
	"fmt"
	set "github.com/deckarep/golang-set"
	biligo "github.com/eric2788/biligo-live"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var (
	listening = set.NewSet()
	excepted  = set.NewSet()
	liveFetch = set.NewSet()
)

func coolDownLiveFetch(room int64) {
	liveFetch.Add(room)
	<-time.After(time.Minute * 5)
	liveFetch.Remove(room)
}

var Debug = true

func LaunchLiveServer(room int64, handle func(data *LiveInfo, msg biligo.Msg)) (context.CancelFunc, error) {

	liveInfo, err := GetLiveInfo(room) // 獲取直播資訊

	if err != nil {

		if _, ok := err.(*TooFastError); ok {
			// 假設為已添加監聽以防止重複監聽
			listening.Add(room)
			go func() {
				log.Printf("將於十分鐘後再嘗試監聽直播: %d", room)
				// 十分鐘冷卻後再重試
				<-time.After(time.Minute * 10)
				listening.Remove(room)
			}()
		}

		return nil, err
	}

	live := biligo.NewLive(Debug, 30*time.Second, 0, func(err error) {
		log.Fatal(err)
	})

	if err := live.Conn(websocket.DefaultDialer, biligo.WsDefaultHost); err != nil {
		log.Println("連接伺服器時出現錯誤: ", err)
		return nil, err
	}

	ctx, stop := context.WithCancel(context.Background())

	go func() {

		if err := live.Enter(ctx, room, "", 0); err != nil {
			log.Printf("監聽房間 %v 時出現錯誤: %v\n", room, err)
			listening.Remove(room)
			stop()
		}

	}()

	go func() {
		for {
			select {
			case tp := <-live.Rev:
				if tp.Error != nil {
					log.Println(tp.Error)
					continue
				}
				// 開播 !?
				if _, ok := tp.Msg.(*biligo.MsgLive); ok {

					// 更新直播資訊只做一次
					if !liveFetch.Contains(room) {
						go coolDownLiveFetch(room)
						log.Printf("房間 %v 開播，正在更新直播資訊...\n", room)
						// 更新一次直播资讯
						UpdateLiveInfo(liveInfo, room)
					}

					// 但開播指令推送多次保留
				}
				handle(liveInfo, tp.Msg)
			case <-ctx.Done():
				log.Printf("房間 %v 監聽中止。\n", room)
				listening.Remove(room)
				return
			}
		}
	}()

	listening.Add(room)
	return stop, nil
}

type LiveInfo struct {
	RoomId int64  `json:"room_id"`
	UID    int64  `json:"uid"`
	Title  string `json:"title"`
	Name   string `json:"name"`
	Cover  string `json:"cover"`
}

type TooFastError struct {
	RoomId int64
}

func (e *TooFastError) Error() string {
	return fmt.Sprintf("房間 %v 請求頻繁", e.RoomId)
}

// UpdateLiveInfo 刷新直播資訊，強制更新緩存
func UpdateLiveInfo(info *LiveInfo, room int64) {

	latestRoomInfo, err := api.GetRoomInfoWithOption(room, true)

	// 房間資訊請求成功
	if err == nil && latestRoomInfo.Code == 0 {
		// 更新房間資訊
		info.Cover = latestRoomInfo.Data.UserCover
		info.Title = latestRoomInfo.Data.Title
		info.UID = latestRoomInfo.Data.Uid
		log.Printf("房間直播資訊 %v 刷新成功。", room)
	} else {
		if err != nil {
			log.Printf("房間直播資訊 %v 刷新失敗: %v", room, err)
		} else {
			log.Printf("房間直播資訊 %v 刷新失敗: %v", room, latestRoomInfo.Message)
		}
	}

	latestUserInfo, err := api.GetUserInfo(info.UID, true)
	// 用戶資訊請求過快被攔截
	if err == nil && latestUserInfo.Code == 0 {
		// 更新用戶資訊
		info.Name = latestUserInfo.Data.Name
		log.Printf("房間用戶資訊 %v 刷新成功。", info.UID)
	} else {
		if err != nil {
			log.Printf("房間用戶資訊 %v 刷新失敗: %v", info.UID, err)
		} else {
			log.Printf("房間用戶資訊 %v 刷新失敗: %v", info.UID, latestUserInfo.Message)
		}
	}
}

// GetLiveInfo 獲取直播資訊，不強制更新緩存
func GetLiveInfo(room int64) (*LiveInfo, error) {

	info, err := api.GetRoomInfoWithOption(room, false)

	if err != nil {
		log.Printf("索取房間資訊 %v 時出現錯誤: %v", room, err)
		return nil, err
	}

	// 房間資訊請求過快被攔截
	if info.Code == -412 {
		log.Printf("錯誤: 房間 %v 請求頻繁被攔截", room)
		return nil, &TooFastError{RoomId: room}
	}

	if info.Data == nil {
		log.Printf("索取房間資訊 %v 時出現錯誤: %v", room, info.Message)
		excepted.Add(room)
		return nil, errors.New(info.Message)
	}

	data := info.Data
	user, err := api.GetUserInfo(data.Uid, false)

	if err != nil {
		log.Println("索取用戶資訊時出現錯誤: ", err)
		return nil, err
	}

	// 用戶資訊請求過快被攔截
	if user.Code == -412 {
		log.Printf("錯誤: 用戶 %v 請求頻繁被攔截", data.Uid)
		return nil, &TooFastError{RoomId: room}
	}

	if user.Data == nil {
		log.Println("索取用戶資訊時出現錯誤: ", user.Message)
		// 404 not found
		if user.Code == -404 {
			log.Printf("用戶 %v 不存在，已排除該房間。", data.Uid)
			excepted.Add(room)
		}
		return nil, errors.New(user.Message)
	}

	liveInfo := &LiveInfo{
		RoomId: room,
		UID:    data.Uid,
		Title:  data.Title,
		Name:   user.Data.Name,
		Cover:  data.UserCover,
	}

	return liveInfo, nil

}
