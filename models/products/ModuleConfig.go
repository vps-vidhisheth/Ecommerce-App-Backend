package products

import (
	"sync"

	"ecommerce/components/log"

	"github.com/jinzhu/gorm"
)

type ModuleConfig struct {
	DB *gorm.DB
}

func NewProductModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{
		DB: db,
	}
}

func (config *ModuleConfig) TableMigration(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	var models []interface{} = []interface{}{ //[]interface{} is like list of anything and here it is
		&Products{}, //pointer to product struct rather than direct struct
	}

	for _, model := range models {
		if err := config.DB.AutoMigrate(model).Error; err != nil {
			log.GetLogger().Print("Auto Migration Error: %s", err)
		}
	}

	log.GetLogger().Print("Products Table Migrated Successfully")
}
