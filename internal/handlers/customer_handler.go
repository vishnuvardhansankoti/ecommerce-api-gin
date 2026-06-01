package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	service services.CustomerService
}

type CustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

func NewCustomerHandler(service services.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

// List godoc
// @Summary List customers
// @Description Get all customers
// @Tags customers
// @Produce json
// @Success 200 {array} models.Customer
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	customers, err := h.service.List(c.Request.Context())
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, customers)
}

// Create godoc
// @Summary Create customer
// @Description Create a new customer
// @Tags customers
// @Accept json
// @Produce json
// @Param request body CustomerRequest true "Customer payload"
// @Success 201 {object} models.Customer
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateCustomerRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer := models.Customer{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Address:   strings.TrimSpace(req.Address),
	}

	createdCustomer, err := h.service.Create(c.Request.Context(), services.CustomerInput{
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Address:   customer.Address,
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, createdCustomer)
}

// Get godoc
// @Summary Get customer
// @Description Get a customer by ID
// @Tags customers
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} models.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/customers/{id} [get]
func (h *CustomerHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.service.Get(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, customer)
}

// Update godoc
// @Summary Update customer
// @Description Update an existing customer
// @Tags customers
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param request body CustomerRequest true "Customer payload"
// @Success 200 {object} models.Customer
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateCustomerRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.service.Update(c.Request.Context(), uint(id), services.CustomerInput{
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Address:   strings.TrimSpace(req.Address),
	})
	if err != nil {
		writeError(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, customer)
}

// Delete godoc
// @Summary Delete customer
// @Description Delete a customer by ID
// @Tags customers
// @Param id path int true "Customer ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/customers/{id} [delete]
func (h *CustomerHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func validateCustomerRequest(req CustomerRequest) error {
	switch {
	case strings.TrimSpace(req.FirstName) == "":
		return errBadRequest("first_name is required")
	case strings.TrimSpace(req.LastName) == "":
		return errBadRequest("last_name is required")
	case strings.TrimSpace(req.Email) == "":
		return errBadRequest("email is required")
	default:
		return nil
	}
}
