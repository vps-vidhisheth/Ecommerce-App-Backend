package db

import (
	mylog "ecommerce/components/log"

	"github.com/jinzhu/gorm"
)

var (
	dbInstance *gorm.DB
)

func SetDB(database *gorm.DB) {
	dbInstance = database
}

func GetDB() *gorm.DB {
	if dbInstance == nil {
		mylog.GetLogger().Print("DB instance not initialized")
	}
	return dbInstance
}
