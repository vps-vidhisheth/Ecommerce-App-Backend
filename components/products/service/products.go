package service

import (
	"fmt"
	"net/url"
	"os"
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

	now := time.Now()
	newProduct.CreatedAt = now
	newProduct.UpdatedAt = now

	for i := range newProduct.Images {
		newProduct.Images[i].ProductID = newProduct.ID
		newProduct.Images[i].CreatedAt = now
		newProduct.Images[i].UpdatedAt = now
	}

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

	productToUpdate.IsActive = productToUpdate.IsActive

	if err := s.repository.Update(uow, productToUpdate); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) DeleteProduct(productToDelete *products.Products) error {
	if _, err := s.doesProductExist(productToDelete.ID); err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	now := time.Now()

	updateMap := map[string]interface{}{
		"DeletedAt": now,
		"is_active": false,
	}
	if err := s.repository.UpdateWithMap(uow, productToDelete, updateMap, repository.Filter("id = ?", productToDelete.ID)); err != nil {
		return err
	}
	updateMapImages := map[string]interface{}{
		"DeletedAt": now,
	}
	if err := s.repository.UpdateWithMap(uow, &products.ProductImage{}, updateMapImages, repository.Filter("product_id = ?", productToDelete.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *ProductService) GetProductByID(id string) (*products.DTO, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var p products.Products
	queryProcessors := []repository.QueryProcessor{
		repository.Filter("id = ?", id),
		repository.NotDeleted(),
		repository.Preload("Images"),
	}

	if err := s.repository.GetRecord(uow, &p, queryProcessors...); err != nil {
		return nil, err
	}

	var images [][]byte
	var imageIDs []uuid.UUID
	for _, img := range p.Images {
		images = append(images, img.Image)
		imageIDs = append(imageIDs, img.ID)
	}

	dto := &products.DTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Images:      images,
		ImageIDs:    imageIDs,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	uow.Commit()
	return dto, nil
}

func (s *ProductService) DeleteProductImage(productID string, imageID uuid.UUID) error {
	if imageID == uuid.Nil {
		return errors.NewValidationError("Invalid image ID")
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var img products.ProductImage
	query := repository.Filter("id = ? AND product_id = ?", imageID, productID)
	if err := s.repository.GetRecord(uow, &img, query); err != nil {
		return errors.NewValidationError("Product image not found")
	}

	updateMap := map[string]interface{}{
		"DeletedAt": time.Now(),
	}
	if err := s.repository.UpdateWithMap(uow, &img, updateMap, repository.Filter("id = ?", imageID)); err != nil {
		return errors.NewValidationError("Failed to delete product image: " + err.Error())
	}

	uow.Commit()
	return nil
}

func (s *ProductService) GetAllProducts(allProducts *[]products.DTO, limit, offset int, totalCount *int, requestForm url.Values) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var productsList []products.Products

	queryProcessors := []repository.QueryProcessor{
		repository.NotDeleted(),
		repository.Filter("is_active = ?", true),
		repository.Paginate(limit, offset, totalCount),
		repository.Preload("Images"),
	}

	if searchQuery := s.buildSearchQuery(requestForm); searchQuery != nil {
		queryProcessors = append(queryProcessors, searchQuery)
	}

	if err := s.repository.GetAll(uow, &productsList, queryProcessors...); err != nil {
		return err
	}

	for _, p := range productsList {
		var images [][]byte
		var imageIDs []uuid.UUID
		for _, img := range p.Images {
			images = append(images, img.Image)
			imageIDs = append(imageIDs, img.ID)
		}

		*allProducts = append(*allProducts, products.DTO{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Images:      images,
			ImageIDs:    imageIDs,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
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
		columns := []string{"name", "description"}
		for _, col := range columns {
			util.AddToSlice(col, "LIKE ?", "OR", "%"+searchTerm+"%", &columnNames, &conditions, &operators, &values)
		}
	}

	return repository.FilterWithOperator(columnNames, conditions, operators, values)
}
func (s *ProductService) BulkCreateProducts(filePath string) error {
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
		if len(row) < 3 {
			return errors.NewValidationError(fmt.Sprintf("Row %d is incomplete", i+1))
		}

		price, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return errors.NewValidationError(fmt.Sprintf("Invalid price at row %d", i+1))
		}

		var images []products.ProductImage
		for col := 3; col < len(row); col++ {
			imgPath := strings.TrimSpace(row[col])
			if imgPath != "" {
				data, err := os.ReadFile(imgPath)
				if err != nil {
					return errors.NewValidationError(fmt.Sprintf("Could not read image %s at row %d: %v", imgPath, i+1, err))
				}
				images = append(images, products.ProductImage{
					Image: data,
				})
			}
		}

		p := products.Products{
			Name:        row[0],
			Description: row[1],
			Price:       price,
			IsActive:    true,
			Images:      images,
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
