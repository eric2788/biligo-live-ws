package api

import (
	"encoding/json"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/database"
	"io"
	"log"
	"net/http"
)

const UserInfoApi = "https://api.bilibili.com/x/space/acc/info?mid=%v&jsonp=jsonp"

func GetUserInfo(uid int64, forceUpdate bool) (*UserInfo, error) {

	dbKey := fmt.Sprintf("user:%v", uid)

	if !forceUpdate {
		var userInfo = &UserInfo{}
		if err := database.GetFromDB(dbKey, userInfo); err == nil {
			return userInfo, nil
		} else {
			if e, ok := err.(*database.EmptyError); ok {
				log.Printf("%v, 正在請求B站 API", e)
			} else {
				log.Printf("從數據庫獲取用戶資訊 %v 時出現錯誤: %v, 正在請求B站 API", uid, err)
			}
		}
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

	if err := database.PutToDB(dbKey, &userInfo); err != nil {
		log.Printf("更新用戶資訊 %v 到數據庫時出現錯誤: %v", uid, err)
	} else {
		log.Printf("更新用戶資訊 %v 到數據庫成功", uid)
	}

	return &userInfo, nil

}

func UserExist(uid int64) (bool, error) {
	res, err := GetUserInfo(uid, false)

	if err != nil {
		return false, err
	}

	return res.Code == 0, nil
}
