package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DbPath = "./cache"
)

var (
	log      = logrus.WithField("service", "database")
	strategy DbStrategy
)

type (
	DbStrategy interface {
		StartDB() error
		CloseDB() error
		GetFromDB(key string, arg interface{}) error
		PutToDB(key string, value interface{}) error
		UpdateDB(update func(db *leveldb.Transaction) error) error
	}

	EmptyError struct {
		Key string
	}
)

func init() {
	t := os.Getenv("DB_STRATEGY")
	switch strings.ToLower(t) {
	case "dynamic":
		strategy = &Dynamic{}
	default:
		strategy = &Singleton{}
	}
}

func (e *EmptyError) Error() string {
	return fmt.Sprintf("Key %v 為空值", e.Key)
}

func StartDB() error {
	return strategy.StartDB()
}

func CloseDB() error {
	return strategy.CloseDB()
}

func closeTransWithLog(tran *leveldb.Transaction) {
	if err := tran.Commit(); err != nil {
		log.Debug("提交事務時出現錯誤:", err)
	}
}

func GetFromDB(key string, arg interface{}) error {
	return strategy.GetFromDB(key, arg)
}

// PutToDB use json to encode value and save into database
func PutToDB(key string, value interface{}) error {
	return strategy.PutToDB(key, value)
}

func UpdateDB(update func(db *leveldb.Transaction) error) error {
	return strategy.UpdateDB(update)
}
