package subscriber

import (
	"fmt"
	"testing"
	"time"
)

// TestGetAllGlobalSubscribers tests the GetAllGlobalSubscribers function concurrently
func TestGetAllGlobalSubscribers(t *testing.T) {

	GetSubscriber("test-1").Global = true
	GetSubscriber("test-2").Global = true
	GetSubscriber("test-3").Global = true

	go func() {
		t.Log("get all subscribers: ")
		<-time.After(time.Second * 2)
		for id := range GetAllGlobalSubscribers() {
			t.Log(id)
		}
	}()

	go func() {
		for i := 4; i < 100; i++ {
			GetSubscriber(fmt.Sprint("test-", i)).Global = true
		}
	}()

	<-time.After(time.Second * 5)
}
