package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockCategoryService struct {
	mock.Mock
}

func (m *mockCategoryService) List(ctx context.Context) ([]models.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *mockCategoryService) Create(ctx context.Context, input services.CategoryInput) (models.Category, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *mockCategoryService) Get(ctx context.Context, id uint) (models.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *mockCategoryService) Update(ctx context.Context, id uint, input services.CategoryInput) (models.Category, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(models.Category), args.Error(1)
}

func (m *mockCategoryService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockProductService struct {
	mock.Mock
}

func (m *mockProductService) List(ctx context.Context, filter services.ProductListFilter) ([]models.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *mockProductService) Create(ctx context.Context, input services.ProductInput) (models.Product, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *mockProductService) Get(ctx context.Context, id uint) (models.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *mockProductService) Update(ctx context.Context, id uint, input services.ProductInput) (models.Product, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *mockProductService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockCustomerService struct {
	mock.Mock
}

func (m *mockCustomerService) List(ctx context.Context) ([]models.Customer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Customer), args.Error(1)
}

func (m *mockCustomerService) Create(ctx context.Context, input services.CustomerInput) (models.Customer, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *mockCustomerService) Get(ctx context.Context, id uint) (models.Customer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *mockCustomerService) Update(ctx context.Context, id uint, input services.CustomerInput) (models.Customer, error) {
	args := m.Called(ctx, id, input)
	return args.Get(0).(models.Customer), args.Error(1)
}

func (m *mockCustomerService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) List(ctx context.Context, filter services.OrderListFilter) ([]models.Order, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *mockOrderService) Create(ctx context.Context, input services.CreateOrderInput) (models.Order, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *mockOrderService) Get(ctx context.Context, id uint) (models.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *mockOrderService) UpdateStatus(ctx context.Context, id uint, status string) (models.Order, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(models.Order), args.Error(1)
}

func TestCategoryCreateReturnsCreatedCategory(t *testing.T) {
	t.Setenv("GIN_MODE", gin.TestMode)
	gin.SetMode(gin.TestMode)

	mockService := &mockCategoryService{}
	handler := NewCategoryHandler(mockService)

	mockService.
		On("Create", mock.Anything, services.CategoryInput{Name: "Electronics", Description: "Devices"}).
		Return(models.Category{BaseModel: models.BaseModel{ID: 1}, Name: "Electronics", Description: "Devices"}, nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":" Electronics ","description":" Devices "}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.POST("/categories", handler.Create)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.JSONEq(t, `{"id":1,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z","name":"Electronics","description":"Devices"}`, recorder.Body.String())
	mockService.AssertExpectations(t)
}

func TestCategoryGetReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockCategoryService{}
	handler := NewCategoryHandler(mockService)

	mockService.
		On("Get", mock.Anything, uint(99)).
		Return(models.Category{}, services.NotFoundError{Resource: "category"}).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/categories/99", nil)
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.GET("/categories/:id", handler.Get)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.JSONEq(t, `{"error":"category not found"}`, recorder.Body.String())
	mockService.AssertExpectations(t)
}

func TestProductListParsesCategoryFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockProductService{}
	handler := NewProductHandler(mockService)

	filterMatcher := mock.MatchedBy(func(filter services.ProductListFilter) bool {
		return filter.CategoryID != nil && *filter.CategoryID == 7
	})

	mockService.
		On("List", mock.Anything, filterMatcher).
		Return([]models.Product{{BaseModel: models.BaseModel{ID: 2}, Name: "Mouse", SKU: "MOUSE-1", Price: 49.99, Stock: 5, CategoryID: 7}}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/products?category_id=7", nil)
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.GET("/products", handler.List)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"sku":"MOUSE-1"`)
	mockService.AssertExpectations(t)
}

func TestProductCreateReturnsValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockProductService{}
	handler := NewProductHandler(mockService)

	mockService.
		On("Create", mock.Anything, services.ProductInput{
			Name:        "Mouse",
			Description: "Wireless",
			SKU:         "MOUSE-1",
			Price:       49.99,
			Stock:       10,
			CategoryID:  999,
		}).
		Return(models.Product{}, services.ValidationError{Message: "category not found"}).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(`{"name":"Mouse","description":"Wireless","sku":"MOUSE-1","price":49.99,"stock":10,"category_id":999}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.POST("/products", handler.Create)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"category not found"}`, recorder.Body.String())
	mockService.AssertExpectations(t)
}

func TestCustomerCreateRejectsMissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockCustomerService{}
	handler := NewCustomerHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString(`{"first_name":"Alex","last_name":"Johnson"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.POST("/customers", handler.Create)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"email is required"}`, recorder.Body.String())
	mockService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCustomerUpdateReturnsUpdatedCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockCustomerService{}
	handler := NewCustomerHandler(mockService)

	mockService.
		On("Update", mock.Anything, uint(5), services.CustomerInput{
			FirstName: "Alex",
			LastName:  "Johnson",
			Email:     "alex@example.com",
			Phone:     "+1-555-111-2222",
			Address:   "101 Market Street",
		}).
		Return(models.Customer{
			BaseModel: models.BaseModel{ID: 5},
			FirstName: "Alex",
			LastName:  "Johnson",
			Email:     "alex@example.com",
			Phone:     "+1-555-111-2222",
			Address:   "101 Market Street",
		}, nil).
		Once()

	req := httptest.NewRequest(http.MethodPut, "/customers/5", bytes.NewBufferString(`{"first_name":"Alex","last_name":"Johnson","email":"alex@example.com","phone":"+1-555-111-2222","address":"101 Market Street"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.PUT("/customers/:id", handler.Update)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"email":"alex@example.com"`)
	mockService.AssertExpectations(t)
}

func TestOrderCreateDelegatesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockOrderService{}
	handler := NewOrderHandler(mockService)

	mockService.
		On("Create", mock.Anything, services.CreateOrderInput{
			CustomerID: 1,
			Items:      []services.CreateOrderItemInput{{ProductID: 2, Quantity: 3}},
		}).
		Return(models.Order{
			BaseModel:   models.BaseModel{ID: 11},
			CustomerID:  1,
			Status:      models.OrderStatusPending,
			TotalAmount: 149.97,
		}, nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"customer_id":1,"items":[{"product_id":2,"quantity":3}]}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.POST("/orders", handler.Create)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"total_amount":149.97`)
	mockService.AssertExpectations(t)
}

func TestOrderUpdateStatusRejectsInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockOrderService{}
	handler := NewOrderHandler(mockService)

	req := httptest.NewRequest(http.MethodPatch, "/orders/4/status", bytes.NewBufferString(`{"status":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.PATCH("/orders/:id/status", handler.UpdateStatus)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.JSONEq(t, `{"error":"status must be one of pending, confirmed, shipped, cancelled"}`, recorder.Body.String())
	mockService.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestOrderListReturnsOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockOrderService{}
	handler := NewOrderHandler(mockService)

	mockService.
		On("List", mock.Anything, services.OrderListFilter{Status: "pending"}).
		Return([]models.Order{{BaseModel: models.BaseModel{ID: 1}, CustomerID: 2, Status: "pending", TotalAmount: 25}}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/orders?status=pending", nil)
	recorder := httptest.NewRecorder()

	router := gin.New()
	router.GET("/orders", handler.List)
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var orders []models.Order
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &orders))
	require.Len(t, orders, 1)
	assert.Equal(t, uint(1), orders[0].ID)
	mockService.AssertExpectations(t)
}
