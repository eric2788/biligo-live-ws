package api

import (
	"fmt"
	"testing"
	"time"
)

const roomId int64 = 545

func TestGetRoomInfo(t *testing.T) {
	if roomInfo, err := GetRoomInfo(roomId); err == nil {
		t.Logf("%#v\n", *roomInfo.Data)
	} else {
		t.Fatal(err)
	}
}

func GetRoomInfoTimer() {
	now := time.Now().UnixMilli()
	defer func() {
		fmt.Printf("Spent: %vms\n", time.Now().UnixMilli()-now)
	}()
	if roomInfo, err := GetRoomInfo(roomId); err == nil {
		fmt.Printf("%#v\n", *roomInfo)
	} else {
		fmt.Println(err)
	}
}

func BenchmarkGetRoomInfo(b *testing.B) {
	for i := 0; i < 10; i++ {
		GetRoomInfoTimer()
	}
}
