package subscriber

import set "github.com/deckarep/golang-set/v2"

var connected = set.NewSet[string]()

// Connected 檢查用戶是否已連線
func AddConnected(identifier string) {
	connected.Add(identifier)
}

func RemoveConnected(identifier string) {
	connected.Remove(identifier)
}