package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type EcommerceRepository interface {
	GetAll(uow *UnitOfWork, out interface{}, queryProcessor ...QueryProcessor) error
	Add(uow *UnitOfWork, out interface{}) error
	UpdateWithMap(uow *UnitOfWork, model interface{}, value map[string]interface{}, queryProcessors ...QueryProcessor) error
	GetRecordForUser(uow *UnitOfWork, userId uuid.UUID, out interface{}, pkColumn string, queryProcessors ...QueryProcessor) error
	Save(uow *UnitOfWork, out interface{}) error
	GetRecord(uow *UnitOfWork, out interface{}, queryProcessors ...QueryProcessor) error
	GetAllInOrder(uow *UnitOfWork, out, orderBy interface{}, queryProcessors ...QueryProcessor) error
	Update(uow *UnitOfWork, out interface{}) error
}

type GormRepository struct{}

func NewGormRespository() *GormRepository {
	return &GormRepository{}
}

type UnitOfWork struct {
	DB        *gorm.DB
	Committed bool
	Readonly  bool
}

func NewUnitOfWork(db *gorm.DB, readonly bool) *UnitOfWork {
	commit := false
	if readonly {
		return &UnitOfWork{
			DB:        db.New(),
			Committed: commit,
			Readonly:  readonly,
		}
	}
	return &UnitOfWork{
		DB:        db.New().Begin(),
		Committed: commit,
		Readonly:  readonly,
	}
}

func (uow *UnitOfWork) Commit() {
	if !uow.Readonly && !uow.Committed {
		uow.Committed = true
		uow.DB.Commit()
	}
}

func (uow *UnitOfWork) RollBack() {
	if !uow.Committed && !uow.Readonly {
		uow.DB.Rollback()
	}
}

func (repository *GormRepository) Add(uow *UnitOfWork, out interface{}) error {
	return uow.DB.Create(out).Error
}

func (repository *GormRepository) Update(uow *UnitOfWork, out interface{}) error {
	return uow.DB.Model(out).Updates(out).Error
}

func (repository *GormRepository) Save(uow *UnitOfWork, value interface{}) error {
	return uow.DB.Save(value).Error
}

func (repository *GormRepository) GetAll(uow *UnitOfWork, out interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB
	db, err := executeQueryProcessors(db, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Find(out).Error
}

func (repository *GormRepository) GetRecord(uow *UnitOfWork, out interface{}, queryProcessor ...QueryProcessor) error {
	db := uow.DB
	db, err := executeQueryProcessors(db, out, queryProcessor...)
	if err != nil {
		return err
	}
	return db.First(out).Error
}

func (repository *GormRepository) GetAllInOrder(uow *UnitOfWork, out, orderBy interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB
	db, err := executeQueryProcessors(db, out, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Order(orderBy).Find(out).Error
}

func (repository *GormRepository) UpdateWithMap(uow *UnitOfWork, model interface{}, value map[string]interface{},
	queryProcessors ...QueryProcessor) error {
	db := uow.DB
	db, err := executeQueryProcessors(db, value, queryProcessors...)
	if err != nil {
		return err
	}
	return db.Model(model).Updates(value).Error
}

type QueryProcessor func(db *gorm.DB, out interface{}) (*gorm.DB, error)

func executeQueryProcessors(db *gorm.DB, out interface{}, queryProcessors ...QueryProcessor) (*gorm.DB, error) {
	var err error
	for _, query := range queryProcessors {
		if query != nil {
			db, err = query(db, out)
			if err != nil {
				return db, err
			}
		}
	}
	return db, err
}

func (repository *GormRepository) Delete(uow *UnitOfWork, out interface{}, queryProcessors ...QueryProcessor) error {
	db := uow.DB
	var err error
	if len(queryProcessors) > 0 {
		db, err = executeQueryProcessors(db, out, queryProcessors...)
		if err != nil {
			return err
		}
	}
	return db.Delete(out).Error
}

func Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args...)
		return db, nil
	}
}

func Select(query interface{}, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Select(query, args...)
		return db, nil
	}
}

func Preload(associations ...string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		for _, assoc := range associations {
			db = db.Preload(assoc)
		}
		return db, nil
	}
}

func Paginate(limit, offset int, totalCount *int) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if totalCount != nil {
			var count int
			if err := db.Model(out).Count(&count).Error; err != nil {
				return db, err
			}
			*totalCount = count
		}
		if limit > 0 {
			db = db.Limit(limit)
		}
		if offset > 0 {
			db = db.Offset(offset)
		}
		db = db.Order("created_at ASC")
		return db, nil
	}
}

func Join(query string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Joins(query, args...)
		return db, nil
	}
}

func Table(tableName string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Table(tableName)
		return db, nil
	}
}

func DoesRecordExistForUser(db *gorm.DB, out interface{}, id uuid.UUID, pkColumn string, queryProcessors ...QueryProcessor) (bool, error) {
	count := 0
	db, err := executeQueryProcessors(db, out, queryProcessors...)
	if err != nil {
		return false, err
	}
	if err := db.Model(out).Where(fmt.Sprintf("%s = ?", pkColumn), id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *GormRepository) GetRecordForUser(uow *UnitOfWork, userId uuid.UUID, out interface{}, pkColumn string, queryProcessors ...QueryProcessor) error {
	queryProcessors = append([]QueryProcessor{Filter(fmt.Sprintf("%s = ?", pkColumn), userId)}, queryProcessors...)
	return repository.GetRecord(uow, out, queryProcessors...)
}

func FilterWithOperator(columnNames []string, conditions []string, operators []string, values []interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {

		if len(columnNames) != len(conditions) && len(conditions) != len(values) {
			return db, nil
		}

		if len(conditions) == 1 {
			if values[0] == nil {
				db = db.Where(fmt.Sprintf("%v %v", columnNames[0], conditions[0]))
				return db, nil
			}
			db = db.Where(fmt.Sprintf("%v %v", columnNames[0], conditions[0]), values[0])
			return db, nil
		}
		if len(columnNames)-1 != len(operators) {
			return db, nil
		}

		str := ""
		nums := []int{}
		for index := 0; index < len(columnNames); index++ {
			if values[index] == nil {
				nums = append(nums, index)
			}
			if index == len(columnNames)-1 {
				str = fmt.Sprintf("%v%v %v", str, columnNames[index], conditions[index])
			} else {
				str = fmt.Sprintf("%v%v %v %v ", str, columnNames[index], conditions[index], operators[index])
			}
		}
		for ind, num := range nums {
			values = append(values[:num], values[num+1:]...)
			for i := ind; i < len(nums); i++ {
				// This is done to adjust indexes because we sliced.
				nums[i] = nums[i] - 1
			}
		}
		db = db.Where(str, values...)
		return db, nil
	}
}

func NotDeleted() QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Where("deleted_at IS NULL"), nil
	}
}
