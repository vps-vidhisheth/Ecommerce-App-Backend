package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/products/service"
	"ecommerce/models/products"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/util"
	"ecommerce/web"
	"io"
	"net/http"
	"strconv"

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

func (c *ProductController) RegisterRoutes(router *mux.Router) {
	productRouter := router.PathPrefix("/products").Subrouter()

	productRouter.Handle("", authmiddleware.AuthMiddleware(http.HandlerFunc(c.GetAllProducts))).Methods(http.MethodGet)
	productRouter.Handle("/{id}", authmiddleware.AuthMiddleware(http.HandlerFunc(c.GetProductByID))).Methods(http.MethodGet)

	adminRouter := router.PathPrefix("/products").Subrouter()
	adminRouter.Handle("", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(c.CreateProduct)))).Methods(http.MethodPost)
	adminRouter.Handle("/{id}", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(c.UpdateProduct)))).Methods(http.MethodPut)
	adminRouter.Handle("/{id}", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(c.DeleteProduct)))).Methods(http.MethodDelete)
	adminRouter.Handle("/bulk", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(c.BulkCreateProducts)))).Methods(http.MethodPost)
	c.log.Print("======== Product Routes Registered =========")
}

func (c *ProductController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	newProduct := products.Products{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		IsActive:    true,
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

	web.RespondJSON(w, http.StatusCreated, newProduct)
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

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64) //form value always returns a string
	productToUpdate := products.Products{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
	}
	productToUpdate.ID = id

	if err := productToUpdate.Validate(true); err != nil {
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

		productToUpdate.Images = append(productToUpdate.Images, products.ProductImage{
			Image: data,
		})
	}

	if err := c.service.UpdateProduct(&productToUpdate); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, productToUpdate)
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
