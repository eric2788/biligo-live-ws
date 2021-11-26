package blive

import (
	"context"
	"errors"
	set "github.com/deckarep/golang-set"
	"github.com/eric2788/biligo-live-ws/services/api"
	"github.com/gorilla/websocket"
	biligo "github.com/iyear/biligo-live"
	"log"
	"time"
)

var listening = set.NewSet()
var excepted = set.NewSet()

func LaunchLiveServer(room int64, handle func(data LiveInfo, msg biligo.Msg)) (context.CancelFunc, error) {

	info, err := api.GetRoomInfo(room)

	if err != nil {
		log.Println("索取房間資訊時出現錯誤: ", err)
		return nil, err
	}

	if info.Data == nil {
		log.Println("索取房間資訊時出現錯誤: ", info.Message)
		return nil, errors.New(info.Message)
	}

	data := info.Data
	realId := data.RoomId // 真正的房間號

	user, err := api.GetUserInfo(data.Uid)

	if err != nil {
		log.Println("索取用戶資訊時出現錯誤: ", err)
		return nil, err
	}

	if user.Data == nil {
		log.Println("索取用戶資訊時出現錯誤: ", err)
		return nil, errors.New(info.Message)
	}

	liveInfo := LiveInfo{
		RoomId:   room,
		UID:      data.Uid,
		Title:    data.Title,
		Name:     user.Name,
		Cover:    data.UserCover,
		RealRoom: realId,
	}

	live := biligo.NewLive(true, 30*time.Second, 0, func(err error) {
		log.Fatal(err)
	})

	if err := live.Conn(websocket.DefaultDialer, biligo.WsDefaultHost); err != nil {
		log.Println("連接伺服器時出現錯誤: ", err)
		return nil, err
	}

	ctx, stop := context.WithCancel(context.Background())

	go func() {
		if err := live.Enter(ctx, realId, "", 0); err != nil {
			log.Println("啟動監聽時出現錯誤: ", err)
			excepted.Add(room)
			listening.Remove(room)
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
	RoomId   int64  `json:"room_id"`
	UID      int64  `json:"uid"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Cover    string `json:"cover"`
	RealRoom int64  `json:"real_room"`
}
