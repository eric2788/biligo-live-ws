package database

import (
	"encoding/json"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Dynamic struct {
	lock sync.Mutex
}

func (d *Dynamic) StartDB() error {
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *Dynamic) CloseDB() error {
	return nil
}

func (d *Dynamic) GetFromDB(key string, arg interface{}) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	db, err := leveldb.OpenFile(DbPath, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		log.Warn("開啟數據庫時出現錯誤:", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Debug("關閉數據庫時出現錯誤:", err)
		}
	}()

	log.Debugf("開始讀取數據庫: %v", key)

	value, err := db.Get([]byte(key), nil)

	if err != nil && err != leveldb.ErrNotFound {
		log.Warn("從數據庫獲取數值時出現錯誤:", err)
		return err
	}

	log.Debugf("讀取數據庫完成: %v", key)

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

func (d *Dynamic) PutToDB(key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		log.Warn("Error encoding value:", err)
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		log.Warn("開啟數據庫時出現錯誤:", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Debug("關閉數據庫時出現錯誤:", err)
		}
	}()
	err = db.Put([]byte(key), b, nil)
	if err != nil {
		log.Warn("更新數據庫時出現錯誤: ", err)
		return err
	}
	return nil
}

func (d *Dynamic) UpdateDB(update func(db *leveldb.Transaction) error) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		log.Warn("開啟數據庫時出現錯誤:", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Debug("關閉數據庫時出現錯誤:", err)
		}
	}()
	tran, err := db.OpenTransaction()
	if err != nil {
		log.Warn("開啟數據庫事務時出現錯誤:", err)
		return err
	}
	defer closeTransWithLog(tran)
	err = update(tran)
	if err != nil {
		log.Warn("更新數據庫時出現錯誤: ", err)
		return err
	}
	return nil
}
