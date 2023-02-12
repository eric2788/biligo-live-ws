package subscriber

import (
	set "github.com/deckarep/golang-set/v2"
	"github.com/gorilla/websocket"
	"sync"
)

type (
	Subscriber struct {
		subscribed set.Set[int64]
		Websocket  *websocket.Conn
		Locker     sync.Mutex
		Global     bool
	}
)

func (s *Subscriber) IsConnected() bool {
	return s.Websocket != nil
}

func (s *Subscriber) UpdateSubscribe(subscribed []int64) {
	s.Locker.Lock()
	defer s.Locker.Unlock()
	s.subscribed = set.NewSet(subscribed...)
}

func (s *Subscriber) GetSubscribed() []int64 {
	return s.subscribed.ToSlice()
}

func (s *Subscriber) AddSubscribes(roomIds []int64) {
	for _, roomId := range roomIds {
		s.subscribed.Add(roomId)
	}
}

func (s *Subscriber) RemoveSubscribes(roomIds []int64) {
	for _, roomId := range roomIds {
		s.subscribed.Remove(roomId)
	}

}

func RemoveSubscriber(clientId string) {
	subscriberMap.Remove(clientId)
}

// GetSubscriber get or create subscriber, so that the subscriber will never be null
// If you wish to check if the subscriber exists, use HasSubscriber instead
func GetSubscriber(clientId string) *Subscriber {
	sub, ok := subscriberMap.Get(clientId)
	if !ok {
		sub = &Subscriber{
			subscribed: set.NewSet[int64](),
		}
		subscriberMap.Set(clientId, sub)
	}
	return sub
}

func HasSubscriber(clientId string) bool {
	return subscriberMap.Has(clientId)
}
