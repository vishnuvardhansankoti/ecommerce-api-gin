package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ecommerce-api-gin/internal/models"
)

type categoryService struct {
	db *sql.DB
}

type productService struct {
	db *sql.DB
}

type customerService struct {
	db *sql.DB
}

type orderService struct {
	db *sql.DB
}

type scanner interface {
	Scan(dest ...any) error
}

func NewCategoryService(db *sql.DB) CategoryService {
	return &categoryService{db: db}
}

func NewProductService(db *sql.DB) ProductService {
	return &productService{db: db}
}

func NewCustomerService(db *sql.DB) CustomerService {
	return &customerService{db: db}
}

func NewOrderService(db *sql.DB) OrderService {
	return &orderService{db: db}
}

func (s *categoryService) List(ctx context.Context) ([]models.Category, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM portaldb.categories
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		category, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *categoryService) Create(ctx context.Context, input CategoryInput) (models.Category, error) {
	category := models.Category{}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO portaldb.categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, input.Name, input.Description).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func (s *categoryService) Get(ctx context.Context, id uint) (models.Category, error) {
	category, err := getCategoryByID(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Category{}, NotFoundError{Resource: "category"}
		}
		return models.Category{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, sku, price, stock, category_id, created_at, updated_at
		FROM portaldb.products
		WHERE category_id = $1
		ORDER BY id ASC
	`, id)
	if err != nil {
		return models.Category{}, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return models.Category{}, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return models.Category{}, err
	}

	category.Products = products
	return category, nil
}

func (s *categoryService) Update(ctx context.Context, id uint, input CategoryInput) (models.Category, error) {
	category := models.Category{}
	result, err := s.db.ExecContext(ctx, `
		UPDATE portaldb.categories
		SET name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3
	`, input.Name, input.Description, id)
	if err != nil {
		return models.Category{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Category{}, err
	}
	if rowsAffected == 0 {
		return models.Category{}, NotFoundError{Resource: "category"}
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM portaldb.categories
		WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Category{}, NotFoundError{Resource: "category"}
		}
		return models.Category{}, err
	}

	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, id uint) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM portaldb.categories WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return NotFoundError{Resource: "category"}
	}

	return nil
}

func (s *productService) List(ctx context.Context, filter ProductListFilter) ([]models.Product, error) {
	query := `
		SELECT p.id, p.name, p.description, p.sku, p.price, p.stock, p.category_id, p.created_at, p.updated_at,
			c.id, c.name, c.description, c.created_at, c.updated_at
		FROM portaldb.products p
		JOIN portaldb.categories c ON c.id = p.category_id
	`
	args := make([]any, 0, 1)
	if filter.CategoryID != nil {
		query += ` WHERE p.category_id = $1`
		args = append(args, *filter.CategoryID)
	}
	query += ` ORDER BY p.id ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		product, err := scanProductWithCategory(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *productService) Create(ctx context.Context, input ProductInput) (models.Product, error) {
	if err := ensureCategoryExists(ctx, s.db, input.CategoryID); err != nil {
		return models.Product{}, err
	}

	inserted := models.Product{}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO portaldb.products (name, description, sku, price, stock, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, sku, price, stock, category_id, created_at, updated_at
	`, input.Name, input.Description, input.SKU, input.Price, input.Stock, input.CategoryID).Scan(
		&inserted.ID,
		&inserted.Name,
		&inserted.Description,
		&inserted.SKU,
		&inserted.Price,
		&inserted.Stock,
		&inserted.CategoryID,
		&inserted.CreatedAt,
		&inserted.UpdatedAt,
	)
	if err != nil {
		return models.Product{}, err
	}

	return s.Get(ctx, inserted.ID)
}

func (s *productService) Get(ctx context.Context, id uint) (models.Product, error) {
	product := models.Product{Category: &models.Category{}}
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.description, p.sku, p.price, p.stock, p.category_id, p.created_at, p.updated_at,
			c.id, c.name, c.description, c.created_at, c.updated_at
		FROM portaldb.products p
		JOIN portaldb.categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.SKU,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Category.ID,
		&product.Category.Name,
		&product.Category.Description,
		&product.Category.CreatedAt,
		&product.Category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Product{}, NotFoundError{Resource: "product"}
		}
		return models.Product{}, err
	}

	return product, nil
}

func (s *productService) Update(ctx context.Context, id uint, input ProductInput) (models.Product, error) {
	if err := ensureCategoryExists(ctx, s.db, input.CategoryID); err != nil {
		return models.Product{}, err
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE portaldb.products
		SET name = $1,
			description = $2,
			sku = $3,
			price = $4,
			stock = $5,
			category_id = $6,
			updated_at = NOW()
		WHERE id = $7
	`, input.Name, input.Description, input.SKU, input.Price, input.Stock, input.CategoryID, id)
	if err != nil {
		return models.Product{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Product{}, err
	}
	if rowsAffected == 0 {
		return models.Product{}, NotFoundError{Resource: "product"}
	}

	return s.Get(ctx, id)
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM portaldb.products WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return NotFoundError{Resource: "product"}
	}

	return nil
}

func (s *customerService) List(ctx context.Context) ([]models.Customer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, first_name, last_name, email, phone, address, created_at, updated_at
		FROM portaldb.customers
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]models.Customer, 0)
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return customers, nil
}

func (s *customerService) Create(ctx context.Context, input CustomerInput) (models.Customer, error) {
	customer := models.Customer{}
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO portaldb.customers (first_name, last_name, email, phone, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, first_name, last_name, email, phone, address, created_at, updated_at
	`, input.FirstName, input.LastName, input.Email, input.Phone, input.Address).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		return models.Customer{}, err
	}

	return customer, nil
}

func (s *customerService) Get(ctx context.Context, id uint) (models.Customer, error) {
	customer := models.Customer{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, email, phone, address, created_at, updated_at
		FROM portaldb.customers
		WHERE id = $1
	`, id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, NotFoundError{Resource: "customer"}
		}
		return models.Customer{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_id, status, total_amount, created_at, updated_at
		FROM portaldb.orders
		WHERE customer_id = $1
		ORDER BY id ASC
	`, id)
	if err != nil {
		return models.Customer{}, err
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		order := models.Order{}
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return models.Customer{}, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return models.Customer{}, err
	}

	customer.Orders = orders
	return customer, nil
}

func (s *customerService) Update(ctx context.Context, id uint, input CustomerInput) (models.Customer, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE portaldb.customers
		SET first_name = $1,
			last_name = $2,
			email = $3,
			phone = $4,
			address = $5,
			updated_at = NOW()
		WHERE id = $6
	`, input.FirstName, input.LastName, input.Email, input.Phone, input.Address, id)
	if err != nil {
		return models.Customer{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Customer{}, err
	}
	if rowsAffected == 0 {
		return models.Customer{}, NotFoundError{Resource: "customer"}
	}

	customer := models.Customer{}
	err = s.db.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, email, phone, address, created_at, updated_at
		FROM portaldb.customers
		WHERE id = $1
	`, id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, NotFoundError{Resource: "customer"}
		}
		return models.Customer{}, err
	}

	return customer, nil
}

func (s *customerService) Delete(ctx context.Context, id uint) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM portaldb.customers WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return NotFoundError{Resource: "customer"}
	}

	return nil
}

func (s *orderService) List(ctx context.Context, filter OrderListFilter) ([]models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.status, o.total_amount, o.created_at, o.updated_at,
			c.id, c.first_name, c.last_name, c.email, c.phone, c.address, c.created_at, c.updated_at
		FROM portaldb.orders o
		JOIN portaldb.customers c ON c.id = o.customer_id
	`
	args := make([]any, 0, 1)
	if strings.TrimSpace(filter.Status) != "" {
		query += ` WHERE o.status = $1`
		args = append(args, filter.Status)
	}
	query += ` ORDER BY o.id ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		order, err := scanOrderWithCustomer(rows)
		if err != nil {
			return nil, err
		}
		items, err := listOrderItems(ctx, s.db, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *orderService) Create(ctx context.Context, input CreateOrderInput) (models.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Order{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var customerExists int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM portaldb.customers WHERE id = $1`, input.CustomerID).Scan(&customerExists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, ValidationError{Message: "customer not found"}
		}
		return models.Order{}, err
	}

	var orderID uint
	err = tx.QueryRowContext(ctx, `
		INSERT INTO portaldb.orders (customer_id, status, total_amount)
		VALUES ($1, $2, 0)
		RETURNING id
	`, input.CustomerID, models.OrderStatusPending).Scan(&orderID)
	if err != nil {
		return models.Order{}, err
	}

	total := 0.0
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return models.Order{}, ValidationError{Message: "product_id is required"}
		}
		if item.Quantity <= 0 {
			return models.Order{}, ValidationError{Message: "quantity must be greater than zero"}
		}

		var productID uint
		var productName string
		var productPrice float64
		var productStock int
		err = tx.QueryRowContext(ctx, `
			SELECT id, name, price, stock
			FROM portaldb.products
			WHERE id = $1
			FOR UPDATE
		`, item.ProductID).Scan(&productID, &productName, &productPrice, &productStock)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return models.Order{}, ValidationError{Message: "product not found"}
			}
			return models.Order{}, err
		}

		if productStock < item.Quantity {
			return models.Order{}, ValidationError{Message: "insufficient stock for product " + productName}
		}

		lineTotal := productPrice * float64(item.Quantity)
		total += lineTotal

		if _, err := tx.ExecContext(ctx, `
			UPDATE portaldb.products
			SET stock = stock - $1,
				updated_at = NOW()
			WHERE id = $2
		`, item.Quantity, productID); err != nil {
			return models.Order{}, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO portaldb.order_items (order_id, product_id, quantity, unit_price, line_total)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, productID, item.Quantity, productPrice, lineTotal); err != nil {
			return models.Order{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE portaldb.orders
		SET total_amount = $1,
			updated_at = NOW()
		WHERE id = $2
	`, total, orderID); err != nil {
		return models.Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Order{}, err
	}

	order, err := getOrderByID(ctx, s.db, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, NotFoundError{Resource: "order"}
		}
		return models.Order{}, err
	}

	return order, nil
}

func (s *orderService) Get(ctx context.Context, id uint) (models.Order, error) {
	order, err := getOrderByID(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, NotFoundError{Resource: "order"}
		}
		return models.Order{}, err
	}

	return order, nil
}

func (s *orderService) UpdateStatus(ctx context.Context, id uint, status string) (models.Order, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE portaldb.orders
		SET status = $1,
			updated_at = NOW()
		WHERE id = $2
	`, status, id)
	if err != nil {
		return models.Order{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Order{}, err
	}
	if rowsAffected == 0 {
		return models.Order{}, NotFoundError{Resource: "order"}
	}

	order, err := getOrderByID(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, NotFoundError{Resource: "order"}
		}
		return models.Order{}, err
	}

	return order, nil
}

func ensureCategoryExists(ctx context.Context, db *sql.DB, categoryID uint) error {
	var id uint
	err := db.QueryRowContext(ctx, `SELECT id FROM portaldb.categories WHERE id = $1`, categoryID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ValidationError{Message: "category not found"}
		}
		return err
	}

	return nil
}

func getCategoryByID(ctx context.Context, db *sql.DB, id uint) (models.Category, error) {
	category := models.Category{}
	err := db.QueryRowContext(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM portaldb.categories
		WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func getOrderByID(ctx context.Context, db *sql.DB, id uint) (models.Order, error) {
	order := models.Order{}
	err := db.QueryRowContext(ctx, `
		SELECT o.id, o.customer_id, o.status, o.total_amount, o.created_at, o.updated_at,
			c.id, c.first_name, c.last_name, c.email, c.phone, c.address, c.created_at, c.updated_at
		FROM portaldb.orders o
		JOIN portaldb.customers c ON c.id = o.customer_id
		WHERE o.id = $1
	`, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.Customer.ID,
		&order.Customer.FirstName,
		&order.Customer.LastName,
		&order.Customer.Email,
		&order.Customer.Phone,
		&order.Customer.Address,
		&order.Customer.CreatedAt,
		&order.Customer.UpdatedAt,
	)
	if err != nil {
		return models.Order{}, err
	}

	items, err := listOrderItems(ctx, db, order.ID)
	if err != nil {
		return models.Order{}, err
	}
	order.Items = items

	return order, nil
}

func listOrderItems(ctx context.Context, db *sql.DB, orderID uint) ([]models.OrderItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.line_total, oi.created_at, oi.updated_at,
			p.id, p.name, p.description, p.sku, p.price, p.stock, p.category_id, p.created_at, p.updated_at
		FROM portaldb.order_items oi
		JOIN portaldb.products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		item := models.OrderItem{}
		product := models.Product{}
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.LineTotal,
			&item.CreatedAt,
			&item.UpdatedAt,
			&product.ID,
			&product.Name,
			&product.Description,
			&product.SKU,
			&product.Price,
			&product.Stock,
			&product.CategoryID,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Product = product
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanCategory(row scanner) (models.Category, error) {
	category := models.Category{}
	if err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	); err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func scanProduct(row scanner) (models.Product, error) {
	product := models.Product{}
	if err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.SKU,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		return models.Product{}, err
	}

	return product, nil
}

func scanCustomer(row scanner) (models.Customer, error) {
	customer := models.Customer{}
	if err := row.Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		return models.Customer{}, err
	}

	return customer, nil
}

func scanProductWithCategory(row scanner) (models.Product, error) {
	product := models.Product{Category: &models.Category{}}
	if err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.SKU,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.Category.ID,
		&product.Category.Name,
		&product.Category.Description,
		&product.Category.CreatedAt,
		&product.Category.UpdatedAt,
	); err != nil {
		return models.Product{}, err
	}

	return product, nil
}

func scanOrderWithCustomer(row scanner) (models.Order, error) {
	order := models.Order{}
	if err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.Customer.ID,
		&order.Customer.FirstName,
		&order.Customer.LastName,
		&order.Customer.Email,
		&order.Customer.Phone,
		&order.Customer.Address,
		&order.Customer.CreatedAt,
		&order.Customer.UpdatedAt,
	); err != nil {
		return models.Order{}, err
	}

	return order, nil
}
