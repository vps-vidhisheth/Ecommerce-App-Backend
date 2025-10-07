package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/products/service"
	Userservice "ecommerce/components/user/service"
	"ecommerce/models/baseStruct"
	"ecommerce/models/products"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/util"
	"ecommerce/web"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductController struct {
	log     log.Log
	service *service.ProductService
}

func NewProductController(productservice *service.ProductService, logger log.Log) *ProductController {
	return &ProductController{
		log:     logger,
		service: productservice,
	}
}

func (c *ProductController) RegisterRoutes(router *mux.Router, userService *Userservice.UserService) {
	productRouter := router.PathPrefix("/products").Subrouter()
	productRouter.Use(func(h http.Handler) http.Handler {
		return authmiddleware.AuthMiddleware(userService, h)
	})

	productRouter.HandleFunc("", c.GetAllProducts).Methods(http.MethodGet)
	productRouter.HandleFunc("/{id}", c.GetProductByID).Methods(http.MethodGet)

	adminRouter := router.PathPrefix("/products").Subrouter()
	adminRouter.Use(func(h http.Handler) http.Handler {
		return authmiddleware.AuthMiddleware(userService, authmiddleware.AdminMiddleware(h))
	})

	adminRouter.HandleFunc("", c.CreateProduct).Methods(http.MethodPost)
	adminRouter.HandleFunc("/{id}", c.UpdateProduct).Methods(http.MethodPut)
	adminRouter.HandleFunc("/{id}", c.DeleteProduct).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/bulk", c.BulkCreateProducts).Methods(http.MethodPost)
	adminRouter.HandleFunc("/{productId}/images/{imageId}", c.DeleteProductImage).Methods(http.MethodDelete)
	c.log.Print("======== Product Routes Registered =========")
}

func (c *ProductController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

	isActive := false
	if strings.ToLower(r.FormValue("is_active")) == "true" {
		isActive = true
	}

	newProduct := products.Products{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		IsActive:    isActive,
	}

	if err := newProduct.Validate(false); err != nil {
		web.RespondError(w, err)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			web.RespondError(w, err)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			web.RespondError(w, err)
			return
		}

		newProduct.Images = append(newProduct.Images, products.ProductImage{
			Image: data,
		})
	}

	if err := c.service.CreateProduct(&newProduct); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, products.ToDTO(&newProduct))

}

func (c *ProductController) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["productId"]
	imageID, err := uuid.Parse(vars["imageId"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.DeleteProductImage(productID, imageID); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, "Product image deleted successfully")
}

func (c *ProductController) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	isActive := false
	if strings.ToLower(r.FormValue("is_active")) == "true" {
		isActive = true
	}

	productToUpdate := products.Products{
		Base:        baseStruct.Base{ID: id},
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		IsActive:    isActive,
	}

	if err := productToUpdate.Validate(true); err != nil {
		web.RespondError(w, err)
		return
	}

	// Add new uploaded images
	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			web.RespondError(w, err)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			web.RespondError(w, err)
			return
		}

		productToUpdate.Images = append(productToUpdate.Images, products.ProductImage{
			Image: data,
		})
	}

	removedImageIds := r.MultipartForm.Value["removedImageIds[]"]
	for _, idStr := range removedImageIds {
		imageUUID, err := uuid.Parse(idStr)
		if err == nil {
			if err := c.service.DeleteProductImage(id.String(), imageUUID); err != nil {
				web.RespondError(w, err)
				return
			}
		}
	}

	if err := c.service.UpdateProduct(&productToUpdate); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, products.ToDTO(&productToUpdate))
}

func (c *ProductController) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	p := products.Products{}
	p.ID = id
	if err := c.service.DeleteProduct(&p); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, "Product deleted successfully")
}

func (c *ProductController) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	p, err := c.service.GetProductByID(id.String())
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, p)
}

func (c *ProductController) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	allProducts := []products.DTO{}
	var totalCount int
	requestForm := r.URL.Query()
	limit, offset := web.NewParser(r).ParseLimitAndOffset()

	if err := c.service.GetAllProducts(&allProducts, limit, offset, &totalCount, requestForm); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allProducts)
}

func (c *ProductController) BulkCreateProducts(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	excelFile, header, err := r.FormFile("file")
	if err != nil {
		web.RespondError(w, err)
		return
	}
	defer excelFile.Close()

	excelFilename, err := util.SaveUploadedFile(excelFile, header, "./uploads/excel", "bulk_products")
	if err != nil {
		web.RespondError(w, err)
		return
	}

	excelPath := "./uploads/excel/" + excelFilename

	if err := c.service.BulkCreateProducts(excelPath); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, "Bulk products created successfully")
}
