package blive

import (
	"errors"
	"github.com/eric2788/biligo-live-ws/services/api"
)

func GetListeningInfo(room int64) (*ListeningInfo, error) {

	liveInfo, err := GetLiveInfoCache(room)
	if err != nil {
		return nil, err
	}

	userInfo, err := api.GetUserInfoCache(liveInfo.UID)

	if err != nil {
		return nil, err
	}

	// 先前沒有記錄
	role := -1

	if &userInfo.Data.Official != nil {
		role = userInfo.Data.Official.Role
	}

	return &ListeningInfo{
		LiveInfo:     liveInfo,
		OfficialRole: role,
	}, nil
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
		log.Debugf("房間直播資訊 %v 刷新成功。", room)
	} else {
		if err != nil {
			log.Warnf("房間直播資訊 %v 刷新失敗: %v", room, err)
		} else {
			log.Warnf("房間直播資訊 %v 刷新失敗: %v", room, latestRoomInfo.Message)
		}
	}

	latestUserInfo, err := api.GetUserInfo(info.UID, true)
	// 用戶資訊請求成功
	if err == nil && latestUserInfo.Code == 0 {
		// 更新用戶資訊
		info.Name = latestUserInfo.Data.Name
		info.UserFace = latestUserInfo.Data.Face
		info.UserDescription = latestUserInfo.Data.Sign

		log.Debugf("房間用戶資訊 %v 刷新成功。", info.UID)
	} else {
		if err != nil {
			log.Warnf("房間用戶資訊 %v 刷新失敗: %v", info.UID, err)
		} else {
			log.Warnf("房間用戶資訊 %v 刷新失敗: %v", info.UID, latestUserInfo.Message)
		}
	}
}

func GetLiveInfoCache(room int64) (*LiveInfo, error) {

	// 已在 exception 內, 則返回不存在
	if excepted.Contains(room) {
		return nil, ErrNotFound
	}

	info, err := api.GetRoomInfoCache(room)

	if err != nil {
		log.Warnf("索取房間資訊 %v 時出現錯誤: %v", room, err)
		return nil, err
	}

	// 房間資訊請求過快被攔截
	if info.Code == -412 {
		log.Warnf("錯誤: 房間 %v 請求頻繁被攔截", room)
		return nil, ErrTooFast
	}

	// 未找到該房間
	if info.Code == 1 {
		log.Warnf("房間不存在 %v", room)
		excepted.Add(room)
		return nil, ErrNotFound
	}

	if info.Data == nil {
		log.Warnf("索取房間資訊 %v 時出現錯誤: %v", room, info.Message)
		excepted.Add(room)
		return nil, errors.New(info.Message)
	}

	data := info.Data
	user, err := api.GetUserInfoCache(data.Uid)

	if err != nil {
		log.Warn("索取用戶資訊時出現錯誤: ", err)
		return nil, err
	}

	// 用戶資訊請求過快被攔截
	if user.Code == -412 {
		log.Warnf("錯誤: 用戶 %v 請求頻繁被攔截", data.Uid)
		return nil, ErrTooFast
	}

	if user.Data == nil {
		log.Warn("索取用戶資訊時出現錯誤: ", user.Message)
		// 404 not found
		if user.Code == -404 {
			log.Warnf("用戶 %v 不存在，已排除該房間。", data.Uid)
			excepted.Add(room)
			return nil, ErrNotFound
		}
		return nil, errors.New(user.Message)
	}

	liveInfo := &LiveInfo{
		RoomId:          info.Data.RoomId,
		UID:             data.Uid,
		Title:           data.Title,
		Name:            user.Data.Name,
		Cover:           data.UserCover,
		UserFace:        user.Data.Face,
		UserDescription: user.Data.Sign,
	}

	return liveInfo, nil
}

// GetLiveInfo 獲取直播資訊，不強制更新緩存
func GetLiveInfo(room int64) (*LiveInfo, error) {

	// 已在 exception 內, 則返回不存在
	if excepted.Contains(room) {
		return nil, ErrNotFound
	}

	info, err := api.GetRoomInfoWithOption(room, false)

	if err != nil {
		log.Warnf("索取房間資訊 %v 時出現錯誤: %v", room, err)
		return nil, err
	}

	// 房間資訊請求過快被攔截
	if info.Code == -412 {
		log.Warnf("錯誤: 房間 %v 請求頻繁被攔截", room)
		return nil, ErrTooFast
	}

	// 未找到該房間
	if info.Code == 1 {
		log.Warnf("房間不存在 %v", room)
		excepted.Add(room)
		return nil, ErrNotFound
	}

	if info.Data == nil {
		log.Warnf("索取房間資訊 %v 時出現錯誤: %v", room, info.Message)
		excepted.Add(room)
		return nil, errors.New(info.Message)
	}

	data := info.Data
	user, err := api.GetUserInfo(data.Uid, false)

	if err != nil {
		log.Warn("索取用戶資訊時出現錯誤: ", err)
		return nil, err
	}

	// 用戶資訊請求過快被攔截
	if user.Code == -412 {
		log.Warnf("錯誤: 用戶 %v 請求頻繁被攔截", data.Uid)
		return nil, ErrTooFast
	}

	if user.Data == nil {
		log.Warn("索取用戶資訊時出現錯誤: ", user.Message)
		// 404 not found
		if user.Code == -404 {
			log.Warnf("用戶 %v 不存在，已排除該房間。", data.Uid)
			excepted.Add(room)
			return nil, ErrNotFound
		}
		return nil, errors.New(user.Message)
	}

	liveInfo := &LiveInfo{
		RoomId:          info.Data.RoomId,
		UID:             data.Uid,
		Title:           data.Title,
		Name:            user.Data.Name,
		Cover:           data.UserCover,
		UserFace:        user.Data.Face,
		UserDescription: user.Data.Sign,
	}

	return liveInfo, nil

}
