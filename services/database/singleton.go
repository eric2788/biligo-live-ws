package database

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

type Singleton struct {
	level *leveldb.DB
}

func (s *Singleton) StartDB() error {
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		return err
	}
	s.level = db
	return nil
}

func (s *Singleton) CloseDB() error {
	return s.level.Close()
}

func (s *Singleton) GetFromDB(key string, arg interface{}) error {
	value, err := s.level.Get([]byte(key), nil)

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

func (s *Singleton) PutToDB(key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Warn("Error encoding value:", err)
		return err
	}
	return s.level.Put([]byte(key), b, nil)
}

func (s *Singleton) UpdateDB(update func(db *leveldb.Transaction) error) error {
	db, err := s.level.OpenTransaction()
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
