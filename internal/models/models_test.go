package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOrderStatusConstants(t *testing.T) {
	statuses := map[string]bool{
		OrderStatusPending:   true,
		OrderStatusConfirmed: true,
		OrderStatusShipped:   true,
		OrderStatusCancelled: true,
	}

	if len(statuses) != 4 {
		t.Fatalf("expected 4 unique order statuses, got %d", len(statuses))
	}
}

func TestCategoryJSONOmitsProductsWhenEmpty(t *testing.T) {
	category := Category{
		Name:        "Electronics",
		Description: "Devices",
	}

	payload, err := json.Marshal(category)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(payload)
	if strings.Contains(jsonStr, "products") {
		t.Fatalf("expected products field to be omitted, got %s", jsonStr)
	}
}

func TestProductJSONIncludesCategoryWhenPresent(t *testing.T) {
	product := Product{
		Name:       "Mouse",
		SKU:        "M-100",
		Price:      49.99,
		Stock:      10,
		CategoryID: 1,
		Category:   &Category{Name: "Electronics", Description: "Devices"},
	}

	payload, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(payload)
	if !strings.Contains(jsonStr, "\"category\"") {
		t.Fatalf("expected category field to be present, got %s", jsonStr)
	}
}
