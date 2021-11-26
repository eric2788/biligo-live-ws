package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	var xResp XResp

	if err := json.Unmarshal(body, &xResp); err != nil {
		return nil, err
	}

	if xResp.Code != 0 {
		return &UserInfo{XResp: xResp}, nil
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	userCaches.Store(uid, &userInfo)
	return &userInfo, nil

}

func UserExist(uid int64) (bool, error) {
	res, err := GetUserInfo(uid)

	if err != nil {
		return false, err
	}

	return res.Code == 0, nil
}
