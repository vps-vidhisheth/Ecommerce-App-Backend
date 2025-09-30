package user

import (
	"sync"
	"time"

	"ecommerce/components/log"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type ModuleConfig struct {
	DB *gorm.DB
}

func NewUserModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{
		DB: db,
	}
}

func (config *ModuleConfig) TableMigration(wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}

	var models []interface{} = []interface{}{
		&User{},
	}

	for _, model := range models {
		if err := config.DB.AutoMigrate(model).Error; err != nil {
			log.GetLogger().Print("Auto Migration Error: %s", err)
		}
	}

	log.GetLogger().Print("User Table Migrated Successfully")

	config.seedAdmin()
}

func (config *ModuleConfig) seedAdmin() {
	db := config.DB
	if db == nil {
		log.GetLogger().Print("DB is nil, cannot seed admin")
		return
	}

	var count int64
	if err := db.Model(&User{}).Where("role = ?", "admin").Count(&count).Error; err != nil {
		log.GetLogger().Print("Failed to count admins:", err)
		return
	}
	if count > 0 {
		log.GetLogger().Print("Admin already exists, skipping seeding.")
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		log.GetLogger().Print("Failed to hash password:", err)
		return
	}

	admin := User{
		ID:        uuid.New(),
		Name:      "Admin",
		Email:     "admin@ecommerce.com",
		Password:  string(pass),
		Role:      "admin",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&admin).Error; err != nil {
		log.GetLogger().Print("Failed to seed admin:", err)
		return
	}

	log.GetLogger().Print("Admin user seeded successfully")
}
