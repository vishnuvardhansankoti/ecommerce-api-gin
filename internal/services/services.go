package services

import (
	"context"
	"errors"
	"strings"

	"ecommerce-api-gin/internal/models"
)

type CategoryInput struct {
	Name        string
	Description string
}

type ProductInput struct {
	Name        string
	Description string
	SKU         string
	Price       float64
	Stock       int
	CategoryID  uint
}

type ProductListFilter struct {
	CategoryID *uint
}

type CustomerInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Address   string
}

type CreateOrderItemInput struct {
	ProductID uint
	Quantity  int
}

type CreateOrderInput struct {
	CustomerID uint
	Items      []CreateOrderItemInput
}

type OrderListFilter struct {
	Status string
}

type CategoryService interface {
	List(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, input CategoryInput) (models.Category, error)
	Get(ctx context.Context, id uint) (models.Category, error)
	Update(ctx context.Context, id uint, input CategoryInput) (models.Category, error)
	Delete(ctx context.Context, id uint) error
}

type ProductService interface {
	List(ctx context.Context, filter ProductListFilter) ([]models.Product, error)
	Create(ctx context.Context, input ProductInput) (models.Product, error)
	Get(ctx context.Context, id uint) (models.Product, error)
	Update(ctx context.Context, id uint, input ProductInput) (models.Product, error)
	Delete(ctx context.Context, id uint) error
}

type CustomerService interface {
	List(ctx context.Context) ([]models.Customer, error)
	Create(ctx context.Context, input CustomerInput) (models.Customer, error)
	Get(ctx context.Context, id uint) (models.Customer, error)
	Update(ctx context.Context, id uint, input CustomerInput) (models.Customer, error)
	Delete(ctx context.Context, id uint) error
}

type OrderService interface {
	List(ctx context.Context, filter OrderListFilter) ([]models.Order, error)
	Create(ctx context.Context, input CreateOrderInput) (models.Order, error)
	Get(ctx context.Context, id uint) (models.Order, error)
	UpdateStatus(ctx context.Context, id uint, status string) (models.Order, error)
}

type NotFoundError struct {
	Resource string
}

func (e NotFoundError) Error() string {
	resource := strings.TrimSpace(e.Resource)
	if resource == "" {
		resource = "resource"
	}

	return resource + " not found"
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return strings.TrimSpace(e.Message)
}

func IsNotFound(err error) bool {
	var target NotFoundError
	return errors.As(err, &target)
}

func IsValidation(err error) bool {
	var target ValidationError
	return errors.As(err, &target)
}
