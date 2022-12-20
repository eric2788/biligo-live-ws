package database

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/kr/pretty"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestGetFromDBConcurrent(t *testing.T) {
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			var v interface{}
			err := GetFromDB(fmt.Sprintf("test:%v", i), &v)
			t.Log(err, v)
		}()
	}

	<-time.After(time.Second * 5)
}

func BenchmarkPutToDB(b *testing.B) {
	b.Log("PutToDB")
	for i := 0; i < b.N; i++ {
		PutToDB(fmt.Sprintf("test:%v", i), i)
	}

	b.Log("GetFromDB")
	for i := 0; i < b.N; i++ {
		var j int
		GetFromDB(fmt.Sprintf("test:%v", i), &j)
	}

}

func TestPutToDBConcurrent(t *testing.T) {

	for i := 0; i < 300; i++ {
		i := i
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			if err != nil {
				t.Log(err)
			}
		}()
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	i := 0

	last := strategy.(*Mix).getStats()

	for {
		<-ticker.C
		i++

		stats := strategy.(*Mix).getStats()
		pretty.Println(stats)

		fmt.Println(strings.Join(pretty.Diff(last, stats), ", "))

		last = stats
		if i == 5 {
			break
		}
	}
}

func TestAtomicAddMinus(t *testing.T) {
	var i atomic.Int64
	i.Add(3)
	assert.Equal(t, i.Load(), int64(3))
	i.Add(-1)
	assert.Equal(t, i.Load(), int64(2))
}

func TestPutToDBAndGetFromDB(t *testing.T) {

	count := 300

	wg := sync.WaitGroup{}

	wg.Add(count * 2)

	for i := 0; i < count; i++ {
		i := i
		go func() {
			var v interface{}
			err := GetFromDB(fmt.Sprintf("test:%v", i), &v)
			if err != nil {
				t.Logf("GetFromDB-%v Error: %v", i, err)
			} else {
				t.Logf("GetFromDB-%v Success: %v", i, v)
			}
			wg.Done()
		}()
	}

	for i := 0; i < count; i++ {
		i := i
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			if err != nil {
				t.Logf("PutToDB-%v Error: %v", i, err)
			} else {
				t.Logf("PutToDB-%v Success", i)
			}
			wg.Done()
		}()
	}

	wg.Wait()

}

func init() {
	strategy = &Mix{}
	logrus.SetLevel(logrus.DebugLevel)
	_ = StartDB()
}
