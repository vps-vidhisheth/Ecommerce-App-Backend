package user

import (
	"ecommerce/components/log"
	"ecommerce/models/baseStruct"
	"ecommerce/repository"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type ModuleConfig struct {
	DB         *gorm.DB
	Repository repository.EcommerceRepository
}

func NewUserModuleConfig(db *gorm.DB) *ModuleConfig {
	return &ModuleConfig{
		DB:         db,
		Repository: repository.NewGormRespository(),
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

	if config.Repository == nil {
		log.GetLogger().Print("Repository is nil, cannot seed admin")
		return
	}

	uow := repository.NewUnitOfWork(db, true)
	defer uow.RollBack()

	var existingAdmin User
	err := config.Repository.GetRecord(uow, &existingAdmin, repository.Filter("role = ?", "admin"))
	if err == nil {
		log.GetLogger().Print("Admin already exists, skipping seeding.")
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		log.GetLogger().Print("Failed to hash password:", err)
		return
	}

	admin := User{
		Base: baseStruct.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:     "Admin",
		Email:    "admin@ecommerce.com",
		Password: string(pass),
		Role:     "admin",
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.GetLogger().Print("Failed to seed admin:", err)
		return
	}

	log.GetLogger().Print("Admin user seeded successfully")
}
