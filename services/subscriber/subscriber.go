package subscriber

import (
	set "github.com/deckarep/golang-set"
	"sync"
)

var subscribeMap = sync.Map{}

func Update(ip string, rooms []int64) {
	subscribeMap.Store(ip, rooms)
}

func Get(ip string) ([]int64, bool) {
	if res, ok := subscribeMap.Load(ip); ok {
		return res.([]int64), ok
	} else {
		return nil, ok
	}
}

func GetOrEmpty(ip string) ([]int64, bool) {
	res, ok := Get(ip)
	if !ok {
		res = []int64{}
	}
	return res, ok
}

func Poll(ip string) ([]int64, bool) {
	if res, ok := subscribeMap.LoadAndDelete(ip); ok {
		return res.([]int64), ok
	} else {
		return nil, ok
	}
}

func GetAllRooms() set.Set {
	rooms := set.NewSet()
	subscribeMap.Range(func(k, v interface{}) bool {
		for _, room := range v.([]int64) {
			rooms.Add(room)
		}
		return true
	})
	return rooms
}

func GetAllSubscribers(room int64) []string {
	ips := make([]string, 0)
	subscribeMap.Range(func(ip, rooms interface{}) bool {
		for _, rm := range rooms.([]int64) {
			if room == rm {
				ips = append(ips, ip.(string))
				break
			}
		}
		return true
	})
	return ips
}

func Delete(ip string) {
	subscribeMap.Delete(ip)
}
