package blive

import (
	mapset "github.com/deckarep/golang-set"
	"log"
	"testing"
	"time"
)

var a = mapset.NewSet(3, 4, 5, 6)
var b = mapset.NewSet(1, 2, 3, 4)

func TestADiffB(t *testing.T) {
	t.Log(a.Difference(b))
}

func TestBDiffA(t *testing.T) {
	t.Log(b.Difference(a))
}

func TestRemoveNonExist(t *testing.T) {
	set := mapset.NewSet(1, 2, 4)
	set.Remove(3)
}

func TestSet(t *testing.T) {
	arr := []int{1, 2, 3, 4}
	for i, k := range mapset.NewSet(&arr).ToSlice() {
		t.Logf("%v: %v", i, k)
	}
}

func TestPanic(t *testing.T) {
	t.Logf("wait 15 seconds")
	c := time.After(time.Second * 15)
	go Rev()
	<-c
	defer func() {
		err := recover()
		t.Logf("recovered: %v", err)
	}()
}

func Rev() {
	log.Printf("panic after 5 seconds")
	<-time.After(time.Second * 5)
	panic("test!")
}
