package httpserver

import (
	"testing"

	"frame_control_system/internal/models"
)

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		from models.OrderStatus
		to   models.OrderStatus
		ok   bool
	}{
		{models.OrderStatusCreated, models.OrderStatusInProgress, true},
		{models.OrderStatusCreated, models.OrderStatusCancelled, true},
		{models.OrderStatusInProgress, models.OrderStatusDone, true},
		{models.OrderStatusInProgress, models.OrderStatusCancelled, true},
		{models.OrderStatusDone, models.OrderStatusCancelled, false},
		{models.OrderStatusCancelled, models.OrderStatusDone, false},
	}
	for _, tt := range tests {
		err := validateTransition(tt.from, tt.to)
		if tt.ok && err != nil {
			t.Fatalf("expected ok from %s to %s, got error %v", tt.from, tt.to, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("expected error from %s to %s", tt.from, tt.to)
		}
	}
}


