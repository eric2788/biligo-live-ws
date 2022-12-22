package api

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eric2788/biligo-live-ws/services/database"
	"github.com/go-ping/ping"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const websocketApi = "https://api.live.bilibili.com/room/v1/Danmu/getConf?room_id=%v&platform=pc&player=web"

func GetWebSocketInfoCache(roomId int64) (*WebSocketInfo, error) {

	dbKey := fmt.Sprintf("wsInfo:%v", roomId)

	var wsInfo = &WebSocketInfo{}
	if err := database.GetFromDB(dbKey, wsInfo); err == nil {
		return wsInfo, nil
	} else {
		if _, ok := err.(*database.EmptyError); ok {
			return nil, ErrCacheNotFound
		} else {
			return nil, err
		}
	}
}

func GetWebSocketInfo(roomId int64, forceUpdate bool) (*WebSocketInfo, error) {

	dbKey := fmt.Sprintf("wsInfo:%v", roomId)

	if !forceUpdate {
		if websocketInfo, err := GetWebSocketInfoCache(roomId); err == nil {
			return websocketInfo, nil
		} else {
			if err == ErrCacheNotFound {
				log.Debugf("%v, 正在請求B站 API", err)
			} else {
				log.Warnf("從數據庫獲取用戶資訊 %v 時出現錯誤: %v, 正在請求B站 API", roomId, err)
			}
		}
	}

	resp, err := getWithAgent(websocketApi, roomId)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var vResp V1Resp

	if err := json.Unmarshal(body, &vResp); err != nil {
		return nil, err
	}

	if vResp.Code != 0 {
		return &WebSocketInfo{V1Resp: vResp}, nil
	}

	var webSocketInfo WebSocketInfo
	if err := json.Unmarshal(body, &webSocketInfo); err != nil {
		return nil, err
	}

	if err := database.PutToDB(dbKey, &webSocketInfo); err != nil {
		log.Warnf("更新 WebSocket 資訊 %v 到數據庫時出現錯誤: %v", roomId, err)
	} else {
		log.Debugf("更新 WebSocket 資訊 %v 到數據庫成功", roomId)
	}

	return &webSocketInfo, nil

}

// UpdateLowLatencyHost 不返回數值以使用 go 懸掛
func UpdateLowLatencyHost(roomId int64) {
	_ = GetLowLatencyHost(roomId, true)
}

// GetLowLatencyHost 返回 最低延遲 Host，但會 blocking
func GetLowLatencyHost(roomId int64, forceUpdate bool) string {

	info, err := GetWebSocketInfo(roomId, forceUpdate)

	if err != nil {
		log.Errorf("嘗試獲取房間 %v 的 WebSocket 資訊時錯誤: %v", roomId, err)
		return ""
	}

	if info.Code != 0 {
		log.Errorf("嘗試獲取房間 %v 的 WebSocket 資訊時錯誤: %v", roomId, info.Msg)
		return ""
	}

	// 已有記錄且不需要強制更新
	if info.LowLatencyHost != "" && !forceUpdate {
		return info.LowLatencyHost
	}

	dbKey := fmt.Sprintf("wsInfo:%v", roomId)

	lowLatencyHost := getLowLatencyHost(info.Data.HostServerList)

	// 不保存
	if lowLatencyHost == "" {
		return ""
	}

	info.LowLatencyHost = fmt.Sprintf("wss://%v/sub", lowLatencyHost)

	if err := database.PutToDB(dbKey, info); err != nil {
		log.Warnf("更新 WebSocket 資訊 %v 到數據庫時出現錯誤: %v", roomId, err)
	} else {
		log.Debugf("更新 WebSocket 資訊 %v 到數據庫成功", roomId)
	}

	return info.LowLatencyHost
}

func getLowLatencyHost(infos []HostServerInfo) string {
	var minPing atomic.Value
	minPing.Store(&LowPingInfo{Ping: 999999999999})
	wg := &sync.WaitGroup{}
	for _, info := range infos {
		wg.Add(1)
		go func(info HostServerInfo) {
			defer wg.Done()
			p, err := ping.NewPinger(info.Host)
			defer p.Stop()
			p.Count = 1
			p.SetPrivileged(true)
			p.Timeout = time.Second * 5
			if err != nil {
				log.Debugf("無法解析 %v :%v", info.Host, err)
				return
			}
			err = p.Run()
			if err != nil {
				log.Debugf("嘗試檢測 %v 的延遲時出現錯誤: %v", info.Host, err)
				return
			}
			stats := p.Statistics()

			log.Debugf("[%v] 已發送封包: %v", info.Host, stats.PacketsSent)
			log.Debugf("[%v] 已接收封包: %v", info.Host, stats.PacketsRecv)
			log.Debugf("[%v] 掉包率: %v", info.Host, stats.PacketLoss)

			// packet loss
			if stats.PacketLoss > 50 {
				log.Debugf("%v 的掉包率大於 50, 已略過", info.Host)
				return
			}

			avgPtt := stats.AvgRtt
			current := minPing.Load().(*LowPingInfo)
			log.Debugf("目前最少延遲: %v (%v)", current.Ping, current.Host)
			log.Debugf("%v 的延遲: %v", info.Host, avgPtt)

			if avgPtt < current.Ping {
				log.Debugf("已成功切換到 %v", info.Host)
				minPing.Store(&LowPingInfo{Host: info.Host, Ping: avgPtt})
			}
		}(info)
	}

	wg.Wait()
	return minPing.Load().(*LowPingInfo).Host
}

type LowPingInfo struct {
	Host string
	Ping time.Duration
}

func ResetAllLowLatency() {
	err := database.UpdateDB(func(db *leveldb.Transaction) error {
		iter := db.NewIterator(&util.Range{Start: []byte("wsInfo:")}, nil)
		defer iter.Release()
		for iter.Next() {
			var wsInfo = &WebSocketInfo{}
			if err := json.Unmarshal(iter.Value(), wsInfo); err != nil {
				log.Errorf("嘗試 decode %v 的數據時錯誤: %v, 已略過", string(iter.Key()), err)
				continue
			}
			// 本身沒有設置
			if wsInfo.LowLatencyHost == "" {
				continue
			}
			wsInfo.LowLatencyHost = ""
			if b, err := json.Marshal(wsInfo); err != nil {
				log.Errorf("嘗試重新 encode %v 的數據時錯誤: %v, 已略過", string(iter.Key()), err)
				continue
			} else {
				if err := db.Put(iter.Key(), b, nil); err != nil {
					log.Errorf("數據 %v 重設失敗: %v", string(iter.Key()), err)
				} else {
					log.Infof("數據 %v 重設成功。", string(iter.Key()))
				}
			}

		}
		return iter.Error()
	})

	if err != nil {
		log.Warnf("重設所有房間的低延遲 Host 時出現錯誤: %v", err)
	} else {
		log.Infof("已重設所有房間的低延遲 Host")
	}
}
