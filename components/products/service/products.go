package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ecommerce/errors"
	"ecommerce/models/products"
	"ecommerce/repository"
	"ecommerce/util"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
)

type ProductService struct {
	db           *gorm.DB
	repository   repository.EcommerceRepository
	associations []string
}

func NewProductService(db *gorm.DB, repo repository.EcommerceRepository, associations []string) *ProductService {
	return &ProductService{
		db:           db,
		repository:   repo,
		associations: associations,
	}
}

func (s *ProductService) CreateProduct(newProduct *products.Products) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	newProduct.ID = uuid.New()
	newProduct.CreatedAt = time.Now()
	newProduct.UpdatedAt = time.Now()
	newProduct.IsActive = true

	if err := s.repository.Add(uow, newProduct); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) UpdateProduct(productToUpdate *products.Products) error {
	existing, err := s.doesProductExist(productToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	productToUpdate.CreatedAt = existing.CreatedAt
	productToUpdate.UpdatedAt = time.Now()

	if err := s.repository.Update(uow, productToUpdate); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) DeleteProduct(productToDelete *products.Products) error {
	_, err := s.doesProductExist(productToDelete.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	now := time.Now()
	updateMap := map[string]interface{}{
		"DeletedAt": now,
		"isActive":  false,
	}

	if err := s.repository.UpdateWithMap(uow, productToDelete, updateMap, repository.Filter("id=?", productToDelete.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) GetProductByID(id uuid.UUID) (*products.Products, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var p products.Products
	if err := s.repository.GetRecord(uow, &p, repository.Filter("id = ?", id)); err != nil {
		return nil, err
	}

	uow.Commit()
	return &p, nil
}

func (s *ProductService) GetAllProducts(allProducts *[]products.DTO, limit, offset int, totalCount *int, requestForm url.Values) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var queryProcessors []repository.QueryProcessor
	searchQuery := s.buildSearchQuery(requestForm)
	if searchQuery != nil {
		queryProcessors = append(queryProcessors, searchQuery)
	}

	queryProcessors = append(queryProcessors, func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Where("deleted_at IS NULL"), nil
	})
	queryProcessors = append(queryProcessors, repository.Paginate(limit, offset, totalCount))

	if err := s.repository.GetAll(uow, allProducts, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) doesProductExist(ID uuid.UUID) (*products.Products, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var p products.Products
	if err := s.repository.GetRecord(uow, &p, repository.Filter("id = ?", ID)); err != nil {
		return nil, errors.NewValidationError("Product ID is invalid")
	}

	uow.Commit()
	return &p, nil
}

func (s *ProductService) buildSearchQuery(requestForm url.Values) repository.QueryProcessor {
	if len(requestForm) == 0 {
		return nil
	}

	var columnNames, conditions, operators []string
	var values []interface{}

	if searchTerm := requestForm.Get("search"); searchTerm != "" {
		columns := []string{"`name`", "`description`"}
		for _, col := range columns {
			util.AddToSlice(col, "LIKE ?", "OR", "%"+searchTerm+"%", &columnNames, &conditions, &operators, &values)
		}
	}

	return repository.FilterWithOperator(columnNames, conditions, operators, values)
}

func (s *ProductService) BulkCreateProducts(filePath string, imagesMap map[string][]byte) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 4 {
			return errors.NewValidationError(fmt.Sprintf("Row %d is incomplete", i+1))
		}

		price, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return errors.NewValidationError(fmt.Sprintf("Invalid price at row %d", i+1))
		}

		var imagesBytes [][]byte
		imageNames := strings.Split(row[3], ",")
		for _, img := range imageNames {
			img = strings.TrimSpace(img)
			if b, ok := imagesMap[img]; ok {
				imagesBytes = append(imagesBytes, b)
			}
		}

		var imgJSON datatypes.JSON
		if len(imagesBytes) > 0 {
			bytes, _ := json.Marshal(imagesBytes)
			imgJSON = datatypes.JSON(bytes)
		}

		p := products.Products{
			ID:          uuid.New(),
			Name:        row[0],
			Description: row[1],
			Price:       price,
			IsActive:    true,
			Images:      imgJSON,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := p.Validate(false); err != nil {
			return errors.NewValidationError(fmt.Sprintf("Row %d validation failed: %v", i+1, err))
		}

		if err := s.repository.Add(uow, &p); err != nil {
			return err
		}
	}

	uow.Commit()
	return nil
}
