package micronova

import (
	"micronova2mqtt/files"
	"testing"
	"time"
)

// MockDataManager creates a mock DataManager for testing
func NewMockDataManager() *files.DataManager {
	pc := files.Power{
		On:  "on",
		Off: "off",
	}
	rk := files.RegKey{
		Key:   "Key",
		Title: "Topic",
	}
	mn := files.Micronova{
		Power:   pc,
		RegKeys: []files.RegKey{rk},
	}

	return &files.DataManager{
		Session: files.Session{
			UUID:         "test-uuid",
			Token:        "test-token",
			RefreshToken: "test-refresh-token",
			ProductId:    "test-product-id",
			DeviceId:     "test-device-id",
		},
		Config: files.Config{
			Micronova: mn,
		},
	}
}

// TestParameterStructure tests parameter data structure
func TestParameterStructure(t *testing.T) {
	param := parameter{
		regKey:  "reg_key_1",
		title:   "topic_1",
		offset:  10,
		mask:    255,
		minimum: 0,
		maximum: 100,
		value:   50,
		formula: "x * 2",
		format:  "%.2f",
		text:    "Test Parameter",
	}

	if param.regKey != "reg_key_1" {
		t.Errorf("Expected regKey 'reg_key_1', got %s", param.regKey)
	}
	if param.value != 50 {
		t.Errorf("Expected value 50, got %d", param.value)
	}
}

// TestConstantsValidity tests that constants have expected values
func TestConstantsValidity(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"durationActive", durationActive, 30 * time.Second},
		{"durationChecking", durationChecking, time.Minute},
		{"maxInactiveCycles", maxInactiveCycles, 180},
		{"stateInactive", stateInactive, 0},
		{"stateActive", stateActive, 1},
		{"stateOffline", stateOffline, 2},
		{"stateNotResponding", stateNotResponding, 3},
	}

	for _, tt := range tests {
		if tt.value != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.value)
		}
	}
}
