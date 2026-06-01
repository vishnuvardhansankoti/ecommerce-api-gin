package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"ecommerce-api-gin/internal/database"
	"ecommerce-api-gin/internal/models"
	"ecommerce-api-gin/internal/router"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	integrationSetupOnce sync.Once
	integrationDB        *sql.DB
	integrationContainer *tcpostgres.PostgresContainer
	integrationSetupErr  error
)

func TestMain(m *testing.M) {
	code := m.Run()

	if integrationContainer != nil {
		_ = integrationContainer.Terminate(context.Background())
	}

	os.Exit(code)
}

func TestCategoryProductCustomerOrderFlow(t *testing.T) {
	testRouter := integrationRouter(t)

	categoryResp := performJSONRequest(t, testRouter, http.MethodPost, "/api/v1/categories", `{"name":"Electronics","description":"Devices"}`)
	assertStatus(t, categoryResp, http.StatusCreated)
	var category models.Category
	decodeJSON(t, categoryResp.Body, &category)

	productResp := performJSONRequest(t, testRouter, http.MethodPost, "/api/v1/products", `{"name":"Wireless Mouse","description":"Bluetooth","sku":"MOUSE-1001","price":49.99,"stock":10,"category_id":1}`)
	assertStatus(t, productResp, http.StatusCreated)
	var product models.Product
	decodeJSON(t, productResp.Body, &product)

	customerResp := performJSONRequest(t, testRouter, http.MethodPost, "/api/v1/customers", `{"first_name":"Alex","last_name":"Johnson","email":"alex@example.com","phone":"+1-555-555-1212","address":"101 Market Street"}`)
	assertStatus(t, customerResp, http.StatusCreated)
	var customer models.Customer
	decodeJSON(t, customerResp.Body, &customer)

	orderResp := performJSONRequest(t, testRouter, http.MethodPost, "/api/v1/orders", `{"customer_id":1,"items":[{"product_id":1,"quantity":2}]}`)
	assertStatus(t, orderResp, http.StatusCreated)
	var order models.Order
	decodeJSON(t, orderResp.Body, &order)

	if order.TotalAmount != 99.98 {
		t.Fatalf("expected total amount 99.98, got %.2f", order.TotalAmount)
	}

	productGetResp := performJSONRequest(t, testRouter, http.MethodGet, "/api/v1/products/1", "")
	assertStatus(t, productGetResp, http.StatusOK)
	var updatedProduct models.Product
	decodeJSON(t, productGetResp.Body, &updatedProduct)

	if updatedProduct.Stock != 8 {
		t.Fatalf("expected stock 8, got %d", updatedProduct.Stock)
	}

	statusResp := performJSONRequest(t, testRouter, http.MethodPatch, "/api/v1/orders/1/status", `{"status":"confirmed"}`)
	assertStatus(t, statusResp, http.StatusOK)

	filteredOrdersResp := performJSONRequest(t, testRouter, http.MethodGet, "/api/v1/orders?status=confirmed", "")
	assertStatus(t, filteredOrdersResp, http.StatusOK)
	var orders []models.Order
	decodeJSON(t, filteredOrdersResp.Body, &orders)

	if len(orders) != 1 || orders[0].Status != models.OrderStatusConfirmed {
		t.Fatalf("expected one confirmed order, got %+v", orders)
	}

	_ = category
	_ = product
	_ = customer
}

func TestProductCreationRequiresExistingCategory(t *testing.T) {
	testRouter := integrationRouter(t)

	response := performJSONRequest(t, testRouter, http.MethodPost, "/api/v1/products", `{"name":"Keyboard","description":"Mechanical","sku":"KEY-2000","price":89.99,"stock":4,"category_id":999}`)
	assertStatus(t, response, http.StatusBadRequest)

	var payload map[string]string
	decodeJSON(t, response.Body, &payload)

	if payload["error"] != "category not found" {
		t.Fatalf("expected category not found error, got %+v", payload)
	}
}

func integrationRouter(t *testing.T) http.Handler {
	t.Helper()

	if !dockerAvailable(t.Context()) {
		t.Skip("docker is not available for testcontainers integration tests")
	}

	integrationSetupOnce.Do(func() {
		integrationContainer, integrationDB, integrationSetupErr = startIntegrationDatabase(t.Context())
	})
	if integrationSetupErr != nil {
		t.Fatalf("integration setup failed: %v", integrationSetupErr)
	}

	resetIntegrationDatabase(t, integrationDB)
	return router.New(integrationDB)
}

func startIntegrationDatabase(ctx context.Context) (*tcpostgres.PostgresContainer, *sql.DB, error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("ecommerce"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, nil, err
	}

	if err := waitForDB(ctx, db, 30*time.Second); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return container, db, nil
}

func waitForDB(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := db.PingContext(waitCtx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-waitCtx.Done():
			if lastErr == nil {
				lastErr = waitCtx.Err()
			}
			return fmt.Errorf("database did not become reachable within %s: %w", timeout, lastErr)
		case <-ticker.C:
		}
	}
}

func resetIntegrationDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	if err := database.ResetSchema(t.Context(), db); err != nil {
		t.Fatalf("failed to reset schema: %v", err)
	}
}

func dockerAvailable(ctx context.Context) bool {
	command := exec.CommandContext(ctx, "docker", "info")
	return command.Run() == nil
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path, payload string) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != "" {
		body = bytes.NewBufferString(payload)
	}

	request := httptest.NewRequest(method, path, body)
	if payload != "" {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Result()
}

func decodeJSON(t *testing.T, body io.ReadCloser, target any) {
	t.Helper()
	defer body.Close()

	if err := json.NewDecoder(body).Decode(target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()

	if response.StatusCode != expected {
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		t.Fatalf("expected status %d, got %d: %s", expected, response.StatusCode, string(body))
	}
}
