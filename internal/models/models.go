package models

import "time"

const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusCancelled = "cancelled"
)

type BaseModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	BaseModel
	Name        string    `gorm:"size:120;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Products    []Product `json:"products,omitempty"`
}

type Product struct {
	BaseModel
	Name        string    `gorm:"size:160;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	SKU         string    `gorm:"size:80;not null;uniqueIndex" json:"sku"`
	Price       float64   `gorm:"type:numeric(10,2);not null" json:"price"`
	Stock       int       `gorm:"not null" json:"stock"`
	CategoryID  uint      `gorm:"not null" json:"category_id"`
	Category    *Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"category,omitempty"`
}

type Customer struct {
	BaseModel
	FirstName string  `gorm:"size:80;not null" json:"first_name"`
	LastName  string  `gorm:"size:80;not null" json:"last_name"`
	Email     string  `gorm:"size:160;not null;uniqueIndex" json:"email"`
	Phone     string  `gorm:"size:30" json:"phone"`
	Address   string  `gorm:"type:text" json:"address"`
	Orders    []Order `json:"orders,omitempty"`
}

type Order struct {
	BaseModel
	CustomerID  uint        `gorm:"not null" json:"customer_id"`
	Customer    Customer    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"customer,omitempty"`
	Status      string      `gorm:"size:32;not null;default:pending" json:"status"`
	TotalAmount float64     `gorm:"type:numeric(10,2);not null" json:"total_amount"`
	Items       []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	BaseModel
	OrderID   uint    `gorm:"index;not null" json:"order_id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Product   Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"product,omitempty"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"type:numeric(10,2);not null" json:"unit_price"`
	LineTotal float64 `gorm:"type:numeric(10,2);not null" json:"line_total"`
}
