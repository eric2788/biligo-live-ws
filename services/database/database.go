package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"sync"
)

const (
	DbPath = "./cache"
)

var (
	lock = sync.Mutex{}
)

type EmptyError struct {
	Key string
}

func (e *EmptyError) Error() string {
	return fmt.Sprintf("Key %v 為空值", e.Key)
}

func StartDB() error {
	_, err := leveldb.OpenFile(DbPath, nil)
	return err
}

func GetFromDB(key string, arg interface{}) error {
	lock.Lock()
	db, err := leveldb.OpenFile(DbPath, nil)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("關閉數據庫時出現錯誤:", err)
		}
		lock.Unlock()
	}()
	if err != nil {
		log.Println("開啟數據庫時出現錯誤:", err)
		return err
	}
	value, err := db.Get([]byte(key), nil)
	if err != nil {
		log.Println("從數據庫獲取數值時出現錯誤:", err)
		return err
	}
	// empty value
	if len(value) == 0 {
		return &EmptyError{key}
	}
	buffer := bytes.NewBuffer(value)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(arg)
	if err != nil {
		log.Println("從數據庫解析數值時出現錯誤:", err)
		return err
	}
	return nil
}

// PutToDB use gob to encode value and save into database
func PutToDB(key string, value interface{}) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(value)
	if err != nil {
		log.Println("Error encoding value:", err)
		return err
	}
	return UpdateDB(func(db *leveldb.DB) error {
		return db.Put([]byte(key), buffer.Bytes(), nil)
	})
}

func UpdateDB(update func(db *leveldb.DB) error) error {
	lock.Lock()
	db, err := leveldb.OpenFile(DbPath, nil)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("關閉數據庫時出現錯誤:", err)
		}
		lock.Unlock()
	}()
	if err != nil {
		log.Println("開啟數據庫時出現錯誤:", err)
		return err
	}
	err = update(db)
	if err != nil {
		log.Println("更新數據庫時出現錯誤: ", err)
		return err
	}
	return nil
}
