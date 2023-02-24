package subscriber

import (
	"time"

	set "github.com/deckarep/golang-set/v2"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sirupsen/logrus"
)

var (
	queue        = set.NewSet[string]()
	expireSet    = set.NewSet[string]()
	subscribeMap = cmap.New[[]int64]()
	log          = logrus.WithField("service", "subscriber")
)

// Update 操作太慢，嘗試使用 go 懸掛
func Update(identifier string, rooms []int64) {
	log.Infof("%v 的訂閱更新已加入隊列...", identifier)
	queue.Add(identifier)
	go func() {
		subscribeMap.Set(identifier, rooms)
		log.Infof("%v 的訂閱更新已完成。", identifier)
		queue.Remove(identifier)
	}()
}

func ExpireAfter(identifier string, timer *time.Timer) {

	if !subscribeMap.Has(identifier) {
		log.Debugf("用戶 %v 未訂閱任何房間, 無需設置過期。", identifier)
		return
	}

	if connected.Contains(identifier) || expireSet.Contains(identifier) {
		log.Debugf("用戶 %v 已連線或已設置過期。", identifier)
		return
	}

	go func() {
		defer timer.Stop()
		<-timer.C

		if connected.Contains(identifier) {
			log.Infof("用戶 %v 已連接WS, 已終止其訂閱過期。", identifier)
			return
		}

		log.Infof("%v 的訂閱已過期。\n", identifier)
		subscribeMap.Remove(identifier)
	}()

	expireSet.Add(identifier)
	log.Infof("已啟動用戶 %v 的訂閱過期。", identifier)
}

func Get(identifier string) ([]int64, bool) {
	if res, ok := subscribeMap.Get(identifier); ok {
		return res, ok
	} else {
		return nil, ok
	}
}

func GetOrEmpty(identifier string) ([]int64, bool) {
	res, ok := Get(identifier)
	if !ok {
		res = []int64{}
	}
	return res, ok
}

func Poll(identifier string) ([]int64, bool) {
	if res, ok := subscribeMap.Pop(identifier); ok {
		return res, ok
	} else {
		return nil, ok
	}
}

func GetAllRooms() set.Set[int64] {
	rooms := set.NewSet[int64]()
	subscribeMap.IterCb(func(key string, value []int64) {
		for _, room := range value {
			rooms.Add(room)
		}
	})
	return rooms
}

func GetAllSubscribers(room int64) []string {
	identifiers := make([]string, 0)
	subscribeMap.IterCb(func(identifier string, rooms []int64) {
		for _, rm := range rooms {
			if room == rm {
				identifiers = append(identifiers, identifier)
				break
			}
		}
	})
	return identifiers
}

func Add(identifier string, rooms []int64) []int64 {

	res, ok := Get(identifier)

	if !ok {
		res = make([]int64, 0)
	}

	newRooms := UpdateRange(res, rooms, func(s set.Set[int64], i int64) {
		s.Add(i)
	})

	Update(identifier, newRooms)
	return newRooms
}

func UpdateRange[T comparable](res []T, rooms []T, updater func(set.Set[T], T)) []T {
	roomSet := ToSet(res)

	for _, room := range rooms {
		updater(roomSet, room)
	}

	return roomSet.ToSlice()
}

func Remove(identifier string, rooms []int64) ([]int64, bool) {

	res, ok := Get(identifier)

	if !ok {
		return nil, false
	}

	newRooms := UpdateRange(res, rooms, func(s set.Set[int64], i int64) {
		s.Remove(i)
	})

	Update(identifier, newRooms)
	return newRooms, true
}

func Delete(identifier string) {
	subscribeMap.Remove(identifier)
}

func ToSet[T comparable](arr []T) set.Set[T] {
	s := set.NewThreadUnsafeSet[T]()
	for _, k := range arr {
		s.Add(k)
	}
	return s
}
