package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eric2788/biligo-live-ws/services/database"
	"io"
	"strings"
	"time"
)

const (
	UserInfoApi    = "https://api.bilibili.com/x/space/acc/info?mid=%v&jsonp=jsonp" // this api url is not valid anymore
	UserInfoApiNew = "https://api.bilibili.com/x/space/wbi/acc/info?mid=%d&platform=web&token=&web_location=1550101&wts=%d&w_rid=%s"
)

var (
	ErrCacheNotFound = errors.New("緩存不存在")
)

func GetUserInfoCache(uid int64) (*UserInfo, error) {

	dbKey := fmt.Sprintf("user:%v", uid)

	var userInfo = &UserInfo{}
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

func wRid(uid, wts int64) string { // 时间戳
	c := "72136226c6a73669787ee4fd02a74c27" // 尾部固定值，根据imgKey,subKey计算得出
	b := "mid=" + fmt.Sprint(uid) + "&platform=web&token=&web_location=1550101"
	a := b + "&wts=" + fmt.Sprint(wts) + c // mid + platform + token + web_location + 时间戳wts + 一个固定值
	hash := md5.Sum([]byte(a))
	return hex.EncodeToString(hash[:])
}

func GetUserInfo(uid int64, forceUpdate bool) (*UserInfo, error) {

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

	timestamp := time.Now().UnixMilli()

	resp, err := getWithAgent(UserInfoApiNew, uid, timestamp, wRid(uid, timestamp))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
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

	userInfo.Data.Face = strings.Replace(userInfo.Data.Face, "http://", "https://", -1)

	if err := database.PutToDB(dbKey, &userInfo); err != nil {
		log.Warnf("更新用戶資訊 %v 到數據庫時出現錯誤: %v", uid, err)
	} else {
		log.Debugf("更新用戶資訊 %v 到數據庫成功", uid)
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
