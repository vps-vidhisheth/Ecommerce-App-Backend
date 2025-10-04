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

	var models []interface{} = []interface{}{
		&Products{},
		&ProductImage{},
	}

	for _, model := range models {
		if err := config.DB.AutoMigrate(model).Error; err != nil {
			log.GetLogger().Print("Auto Migration Error: %s", err)
		}
	}

	log.GetLogger().Print("Products and ProductImages Tables Migrated Successfully")
}
