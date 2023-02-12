package subscriber

import (
	set "github.com/deckarep/golang-set/v2"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	subscriberMap = cmap.New[*Subscriber]() // concurrent thread safe map
	expireSet     = set.NewSet[string]()
	logger        = logrus.WithField("service", "subscriber")
)

func GetSubscribes(clientId string) ([]int64, bool) {
	subscriber, ok := subscriberMap.Get(clientId)
	if !ok {
		return nil, false
	}
	return subscriber.GetSubscribed(), true
}

func GetAllSubscribers(room int64) map[string]*Subscriber {
	var roomSubMap = make(map[string]*Subscriber)
	for id, subscriber := range subscriberMap.Items() {
		if !subscriber.Global && subscriber.subscribed.Contains(room) {
			roomSubMap[id] = subscriber
		}
	}
	return roomSubMap
}

func GetAllGlobalSubscribers() map[string]*Subscriber {
	var roomSubMap = make(map[string]*Subscriber)
	for id, subscriber := range subscriberMap.Items() {
		if subscriber.Global {
			roomSubMap[id] = subscriber
		}
	}
	return roomSubMap
}

// ActiveExpire check if the subscriber is connected, if not, remove it from the subscriberMap
// must use goroutine to call this function or it will block the main thread
func ActiveExpire(clientId string, duration time.Duration) {

	subscriber, ok := subscriberMap.Get(clientId)
	if !ok {
		logger.Debugf("clientId %v 不存在", clientId)
		return
	}

	// 如果已经在过期队列中，或者已经连接上了，就不需要再次过期了
	if subscriber.IsConnected() || expireSet.Contains(clientId) {
		logger.Debugf("clientId %v 已经在过期队列中，或者已经连接上了，就不需要再次过期了", clientId)
		return
	}

	expireSet.Add(clientId)

	timer := time.NewTimer(duration)

	defer timer.Stop()
	defer expireSet.Remove(clientId)
	if subscriber.IsConnected() {
		logger.Infof("%v 已連接，已終止其訂閱過期。", clientId)
		return
	}

	logger.Infof("%v 的訂閱已過期。", clientId)
	RemoveSubscriber(clientId)
}

func GetAllRooms() set.Set[int64] {
	all := set.NewSet[int64]()
	for _, subscriber := range subscriberMap.Items() {
		all = all.Union(subscriber.subscribed)
	}
	return all
}
