package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	service services.ProductService
}

type ProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SKU         string  `json:"sku"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  uint    `json:"category_id"`
}

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// List godoc
// @Summary List products
// @Description Get all products, optionally filtered by category_id
// @Tags products
// @Produce json
// @Param category_id query int false "Category ID"
// @Success 200 {array} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) List(c *gin.Context) {
	filter := services.ProductListFilter{}
	if categoryIDRaw := strings.TrimSpace(c.Query("category_id")); categoryIDRaw != "" {
		categoryID, err := strconv.Atoi(categoryIDRaw)
		if err != nil || categoryID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		parsedCategoryID := uint(categoryID)
		filter.CategoryID = &parsedCategoryID
	}

	products, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, products)
}

// Create godoc
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param request body ProductRequest true "Product payload"
// @Success 201 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateProductRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		SKU:         strings.TrimSpace(req.SKU),
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
	}

	createdProduct, err := h.service.Create(c.Request.Context(), services.ProductInput{
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Price:       product.Price,
		Stock:       product.Stock,
		CategoryID:  product.CategoryID,
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, createdProduct)
}

// Get godoc
// @Summary Get product
// @Description Get a product by ID
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.service.Get(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, product)
}

// Update godoc
// @Summary Update product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body ProductRequest true "Product payload"
// @Success 200 {object} models.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateProductRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.service.Update(c.Request.Context(), uint(id), services.ProductInput{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		SKU:         strings.TrimSpace(req.SKU),
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, product)
}

// Delete godoc
// @Summary Delete product
// @Description Delete a product by ID
// @Tags products
// @Param id path int true "Product ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func validateProductRequest(req ProductRequest) error {
	switch {
	case strings.TrimSpace(req.Name) == "":
		return errBadRequest("name is required")
	case strings.TrimSpace(req.SKU) == "":
		return errBadRequest("sku is required")
	case req.Price <= 0:
		return errBadRequest("price must be greater than zero")
	case req.Stock < 0:
		return errBadRequest("stock cannot be negative")
	case req.CategoryID == 0:
		return errBadRequest("category_id is required")
	default:
		return nil
	}
}
