package database

import (
	"fmt"
	"sync"
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

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			if err != nil {
				t.Log(err)
			}
		}()
	}

	<-time.After(time.Second * 5)
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
			}
			wg.Done()
		}()
	}

	wg.Wait()

}

func init() {
	strategy = &Dynamic{}
	logrus.SetLevel(logrus.DebugLevel)
	_ = StartDB()
}
