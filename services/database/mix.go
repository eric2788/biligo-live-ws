package database

import (
	"context"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"sync/atomic"
	"time"
)

type Mix struct {
	level *leveldb.DB
	mu    sync.Mutex
	ctx   context.Context
	stop  context.CancelFunc
	alive atomic.Int64
}

func (m *Mix) StartDB() error {
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		return err
	}
	return db.Close()
}

func (m *Mix) CloseDB() error {
	return m.closeDB()
}

func (m *Mix) GetFromDB(key string, arg interface{}) error {
	if err := m.initDB(); err != nil {
		return err
	}

	m.alive.Add(1)
	defer m.removeAlive()

	value, err := m.level.Get([]byte(key), nil)

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

func (m *Mix) PutToDB(key string, value interface{}) error {
	if err := m.initDB(); err != nil {
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		log.Warn("Error encoding value:", err)
		return err
	}
	m.alive.Add(1)
	defer m.removeAlive()
	return m.level.Put([]byte(key), b, nil)
}

func (m *Mix) UpdateDB(update func(db *leveldb.Transaction) error) error {
	if err := m.initDB(); err != nil {
		return err
	}
	db, err := m.level.OpenTransaction()
	if err != nil {
		log.Warn("開啟 transaction 時出現錯誤:", err)
		return err
	}

	m.alive.Add(1)
	defer m.removeAlive()
	defer closeTransWithLog(db)

	err = update(db)
	if err != nil {
		log.Warn("更新數據庫時出現錯誤: ", err)
		return err
	}
	return nil
}

func (m *Mix) removeAlive() {
	go func() {
		<-time.After(time.Second)
		m.alive.Add(-1)
	}()
}

func (m *Mix) initDB() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.level != nil {
		return nil
	}
	db, err := leveldb.OpenFile(DbPath, nil)
	if err != nil {
		return err
	}
	m.level = db
	m.ctx, m.stop = context.WithCancel(context.Background())
	go m.checkClose()
	log.Info("數據庫已啟動")
	return nil
}

func (m *Mix) closeDB() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.level == nil {
		return nil
	}
	m.stop()
	err := m.level.Close()
	m.level = nil
	log.Info("數據庫已關閉")
	return err
}

func (m *Mix) getStats() *leveldb.DBStats {
	if m.level == nil {
		return nil
	}
	var stats leveldb.DBStats
	err := m.level.Stats(&stats)
	if err != nil {
		log.Warn("獲取數據庫狀態時出現錯誤:", err)
		return nil
	}
	return &stats
}

func (m *Mix) checkClose() {
	// close the db if there are no more transactions after 1 minute
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if m.level != nil && m.alive.Load() == 0 {
				log.Info("數據庫進入閒置狀態，正在關閉數據庫...")
				err := m.closeDB()
				if err != nil {
					log.Warn("關閉數據庫時出現錯誤:", err)
				}
				return
			}
		case <-m.ctx.Done():
			return
		}
	}
}
