package database

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestGetFromDBConcurrent(t *testing.T) {

	for i := 0; i < 5; i++ {
		go func() {
			var v interface{}
			err := GetFromDB(fmt.Sprintf("test:%v", i), &v)
			log.Println(err, v)
		}()
	}

	<-time.After(time.Second * 5)
}

func TestPutToDBConcurrent(t *testing.T) {

	for i := 0; i < 10; i++ {
		go func() {
			err := PutToDB(fmt.Sprintf("test:%v", i), i)
			log.Println(err)
		}()
	}

	<-time.After(time.Second * 5)
}
