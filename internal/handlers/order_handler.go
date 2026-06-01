package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service services.OrderService
}

type CreateOrderRequest struct {
	CustomerID uint               `json:"customer_id"`
	Items      []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type OrderStatusRequest struct {
	Status string `json:"status"`
}

func NewOrderHandler(service services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// List godoc
// @Summary List orders
// @Description Get all orders, optionally filtered by status
// @Tags orders
// @Produce json
// @Param status query string false "Order status"
// @Success 200 {array} models.Order
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	orders, err := h.service.List(c.Request.Context(), services.OrderListFilter{
		Status: normalizeStatus(c.Query("status")),
	})
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// Create godoc
// @Summary Create order
// @Description Create a new order and decrement product stock
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order payload"
// @Success 201 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CustomerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one item is required"})
		return
	}

	orderItems := make([]services.CreateOrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		orderItems = append(orderItems, services.CreateOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	createdOrder, err := h.service.Create(c.Request.Context(), services.CreateOrderInput{
		CustomerID: req.CustomerID,
		Items:      orderItems,
	})
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, createdOrder)
}

// Get godoc
// @Summary Get order
// @Description Get an order by ID
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.service.Get(c.Request.Context(), uint(id))
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateStatus godoc
// @Summary Update order status
// @Description Update the status of an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body OrderStatusRequest true "Order status payload"
// @Success 200 {object} models.Order
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req OrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := normalizeStatus(req.Status)
	if !isValidOrderStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be one of pending, confirmed, shipped, cancelled"})
		return
	}

	order, err := h.service.UpdateStatus(c.Request.Context(), uint(id), status)
	if err != nil {
		writeError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, order)
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func isValidOrderStatus(status string) bool {
	switch status {
	case models.OrderStatusPending, models.OrderStatusConfirmed, models.OrderStatusShipped, models.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
