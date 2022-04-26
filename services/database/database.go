package database

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DbPath = "./cache"
)

var (
	log   = logrus.WithField("service", "database")
	level *leveldb.DB
)

type EmptyError struct {
	Key string
}

func (e *EmptyError) Error() string {
	return fmt.Sprintf("Key %v 為空值", e.Key)
}

func StartDB() error {
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		return err
	}
	level = db
	return nil
}

func CloseDB() error {
	return level.Close()
}

func closeTransWithLog(tran *leveldb.Transaction) {
	if err := tran.Commit(); err != nil {
		log.Debug("提交事務時出現錯誤:", err)
	}
}

func GetFromDB(key string, arg interface{}) error {
	transaction, err := level.OpenTransaction()
	if err != nil {
		log.Warn("開啟 transaction 時出現錯誤:", err)
		return err
	}

	defer closeTransWithLog(transaction)

	value, err := transaction.Get([]byte(key), nil)

	if err != nil && err != leveldb.ErrNotFound {
		log.Warn("從數據庫獲取數值時出現錯誤:", err)
		return err
	}

	// empty value
	if err == leveldb.ErrNotFound || value == nil || len(value) == 0 {
		return &EmptyError{key}
	}
	err = json.Unmarshal(value, arg)
	if err != nil {
		log.Warn("從數據庫解析數值時出現錯誤:", err)
		return err
	}
	return nil
}

// PutToDB use json to encode value and save into database
func PutToDB(key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Warn("Error encoding value:", err)
		return err
	}

	trans, err := level.OpenTransaction()

	if err != nil {
		log.Warn("開啟 transaction 時出現錯誤:", err)
		return err
	}

	defer closeTransWithLog(trans)

	return trans.Put([]byte(key), b, nil)
}

func UpdateDB(update func(db *leveldb.Transaction) error) error {
	db, err := level.OpenTransaction()
	if err != nil {
		log.Warn("開啟 transaction 時出現錯誤:", err)
		return err
	}

	defer closeTransWithLog(db)

	err = update(db)
	if err != nil {
		log.Warn("更新數據庫時出現錯誤: ", err)
		return err
	}
	return nil
}
