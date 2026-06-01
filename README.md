# E-commerce API with Gin and PostgreSQL

Sample REST API for an e-commerce portal built with:

- [Gin](https://github.com/gin-gonic/gin) for HTTP routing
- Direct SQL queries via Go's `database/sql`
- PostgreSQL as the database
- [Swaggo](https://github.com/swaggo/swag) for Swagger/OpenAPI documentation

## Endpoints

### Health

- `GET /health`
- `GET /swagger/index.html`

### Categories

- `GET /api/v1/categories`
- `POST /api/v1/categories`
- `GET /api/v1/categories/:id`
- `PUT /api/v1/categories/:id`
- `DELETE /api/v1/categories/:id`

### Products

- `GET /api/v1/products`
- `POST /api/v1/products`
- `GET /api/v1/products/:id`
- `PUT /api/v1/products/:id`
- `DELETE /api/v1/products/:id`

### Customers

- `GET /api/v1/customers`
- `POST /api/v1/customers`
- `GET /api/v1/customers/:id`
- `PUT /api/v1/customers/:id`
- `DELETE /api/v1/customers/:id`

### Orders

- `GET /api/v1/orders`
- `POST /api/v1/orders`
- `GET /api/v1/orders/:id`
- `PATCH /api/v1/orders/:id/status`

## Run locally

1. Create a PostgreSQL database named `ecommerce`.
2. Copy `.env.example` values into your environment.
3. Install dependencies and start the server:

```bash
go mod tidy
go run .
```

The application applies schema migrations on startup.

## Database scripts

SQL scripts are available under [scripts](scripts):

- [scripts/001_init_schema.sql](scripts/001_init_schema.sql): creates tables and indexes
- [scripts/002_drop_schema.sql](scripts/002_drop_schema.sql): drops all schema objects used by the API

The API creates and uses the PostgreSQL schema `portaldb` inside the configured database.

Run manually with `psql` if needed:

```bash
psql "$DATABASE_URL" -f scripts/001_init_schema.sql
```

Generate or refresh Swagger docs:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
$(go env GOPATH)/bin/swag init --parseInternal -g main.go -o docs
```

Open Swagger UI at `http://localhost:8080/swagger/index.html`.

## Testing

Run only the unit tests:

```bash
go test -v ./internal/handlers/...
```

Run only the integration tests:

```bash
go test -v .
```

### Integration Test Notes

Integration tests are defined in [api_integration_test.go](api_integration_test.go) and run against a real PostgreSQL container.

Prerequisites:

- Docker Desktop (or Docker Engine) is installed.
- Docker daemon is running before you execute tests.

How they work:

- Tests start a temporary `postgres:16-alpine` container via `testcontainers-go`.
- The test setup resets the schema and recreates all tables under `portaldb`.
- HTTP endpoints are exercised end-to-end using the Gin router.

Useful commands:

```bash
# Run integration tests only
go test -v .

# Run one integration test
go test -v . -run TestCategoryProductCustomerOrderFlow
```

If integration tests fail with container startup errors:

- Verify Docker is running: `docker info`
- Retry after pulling the image once: `docker pull postgres:16-alpine`
- Re-run tests: `go test -v .`

Run the full suite:

```bash
go test -v ./...
```

The integration tests use `testcontainers-go`, so Docker must be installed and running.

## Example payloads

### Create category

```json
{
  "name": "Electronics",
  "description": "Phones, laptops, and accessories"
}
```

### Create product

```json
{
  "name": "Wireless Mouse",
  "description": "Ergonomic mouse with Bluetooth support",
  "sku": "MOUSE-1001",
  "price": 49.99,
  "stock": 120,
  "category_id": 1
}
```

### Create customer

```json
{
  "first_name": "Alex",
  "last_name": "Johnson",
  "email": "alex@example.com",
  "phone": "+1-555-555-1212",
  "address": "101 Market Street, Austin, TX"
}
```

### Create order

```json
{
  "customer_id": 1,
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ]
}
```
