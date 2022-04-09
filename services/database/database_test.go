package database

import (
	"fmt"
	"testing"
	"time"
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

func TestPutToDBConcurrent(t *testing.T) {

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			t.Log(err)
		}()
	}

	<-time.After(time.Second * 5)
}

func TestPutToDBAndGetFromDB(t *testing.T) {
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			var v interface{}
			err := GetFromDB(fmt.Sprintf("test:%v", i), &v)
			t.Log(err, v)
		}()
	}

	for i := 0; i < 10; i++ {
		i := i
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			t.Log(err)
		}()
	}

}

func init() {
	_ = StartDB()
}
