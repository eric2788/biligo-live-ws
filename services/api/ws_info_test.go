package api

import (
	"github.com/kr/pretty"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestGetWebSocketInfo(t *testing.T) {
	info, err := GetWebSocketInfo(8725120, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pretty.Print(info.Data); err != nil {
		t.Fatal(err)
	}
}

func TestGetLowLatencyHost(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	host := GetLowLatencyHost(8725120, false)
	if host == "" {
		t.Error("host is empty")
	} else {
		t.Log("low latency host:", host)
	}
}
