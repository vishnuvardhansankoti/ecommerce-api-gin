package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service services.CategoryService
}

type CategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// List godoc
// @Summary List categories
// @Description Get all product categories
// @Tags categories
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.List(c.Request.Context())
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// Create godoc
// @Summary Create category
// @Description Create a new product category
// @Tags categories
// @Accept json
// @Produce json
// @Param request body CategoryRequest true "Category payload"
// @Success 201 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	category := models.Category{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}

	createdCategory, err := h.service.Create(c.Request.Context(), services.CategoryInput{
		Name:        category.Name,
		Description: category.Description,
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// Get godoc
// @Summary Get category
// @Description Get a category by ID
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := h.service.Get(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, category)
}

// Update godoc
// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body CategoryRequest true "Category payload"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	category, err := h.service.Update(c.Request.Context(), uint(id), services.CategoryInput{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, category)
}

// Delete godoc
// @Summary Delete category
// @Description Delete a category by ID
// @Tags categories
// @Param id path int true "Category ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}
