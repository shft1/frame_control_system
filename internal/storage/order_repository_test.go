package storage

import (
	"testing"

	"frame_control_system/internal/models"
)

func TestCalculateTotal(t *testing.T) {
	items := []models.OrderItem{
		{Name: "a", Quantity: 2, Price: 10},
		{Name: "b", Quantity: 1, Price: 5.5},
	}
	total, err := CalculateTotal(items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 25.5 {
		t.Fatalf("want 25.5, got %v", total)
	}
}

func TestCalculateTotal_Invalid(t *testing.T) {
	items := []models.OrderItem{
		{Name: "a", Quantity: 0, Price: 10},
	}
	_, err := CalculateTotal(items)
	if err == nil {
		t.Fatalf("expected error for invalid item")
	}
}


