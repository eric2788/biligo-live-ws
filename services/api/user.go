package api

import (
	"sync"
)

const UserInfoApi = "https://api.bilibili.com/x/space/acc/info?mid=%v&jsonp=jsonp"

var userCaches = sync.Map{}

func GetUserInfo(uid int64) (*UserInfo, error) {
	if res, ok := userCaches.Load(uid); ok {
		return res.(*UserInfo), nil
	}

	resp, err := http.Get(fmt.Sprintf(UserInfoApi, uid))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var xresp XResp

	if err := json.Unmarshal(body, &xresp); err != nil {
		return nil, err
	}

	if v1resp.Code != 0 {
		return &UserInfo{XResp: xresp}, nil
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	userCaches.Store(uid, &userInfo)
	return &userInfo, nil

}

func UserExist(uid int64) (bool, error) {
	info, err := GetRoomInfo(uid)

	if err != nil {
		return false, err
	}

	return res.Code == 0, nil
}
