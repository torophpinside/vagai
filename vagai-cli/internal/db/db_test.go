package db

import (
	"testing"
)

func TestCheckSitesLimit_ZeroOrgID(t *testing.T) {
	allowed, current, maxLimit, err := CheckSitesLimit(0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !allowed {
		t.Error("Expected allowed = true for orgID=0")
	}
	if current != 0 {
		t.Errorf("Expected current=0, got %d", current)
	}
	if maxLimit != 0 {
		t.Errorf("Expected maxLimit=0, got %d", maxLimit)
	}
}
