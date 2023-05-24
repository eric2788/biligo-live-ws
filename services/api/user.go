package api

import (
	"errors"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/eric2788/common-services/bilibili"
	"strings"
)

var (
	ErrCacheNotFound = errors.New("緩存不存在")
)

func GetUserInfoCache(uid int64) (*bilibili.UserInfo, error) {

	dbKey := fmt.Sprintf("user:%v", uid)

	var userInfo = &bilibili.UserInfo{}
	if err := database.GetFromDB(dbKey, userInfo); err == nil {
		return userInfo, nil
	} else {
		if _, ok := err.(*database.EmptyError); ok {
			return nil, ErrCacheNotFound
		} else {
			return nil, err
		}
	}
}

func GetUserInfo(uid int64, forceUpdate bool) (*bilibili.UserInfo, error) {

	dbKey := fmt.Sprintf("user:%v", uid)

	if !forceUpdate {
		if userInfo, err := GetUserInfoCache(uid); err == nil {
			return userInfo, nil
		} else {
			if err == ErrCacheNotFound {
				log.Debugf("%v, 正在請求B站 API", err)
			} else {
				log.Warnf("從數據庫獲取用戶資訊 %v 時出現錯誤: %v, 正在請求B站 API", uid, err)
			}
		}
	}

	userInfo, err := bilibili.GetUserInfo(uid)
	if err != nil {
		return nil, err
	}

	userInfo.Data.Face = strings.Replace(userInfo.Data.Face, "http://", "https://", -1)

	if err := database.PutToDB(dbKey, &userInfo); err != nil {
		log.Warnf("更新用戶資訊 %v 到數據庫時出現錯誤: %v", uid, err)
	} else {
		log.Debugf("更新用戶資訊 %v 到數據庫成功", uid)
	}

	return userInfo, nil

}

func UserExist(uid int64) (bool, error) {
	res, err := GetUserInfo(uid, false)

	if err != nil {
		return false, err
	}

	return res.Code == 0, nil
}
