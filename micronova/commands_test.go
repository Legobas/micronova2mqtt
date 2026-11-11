package micronova

import (
	"micronova2mqtt/files"
	"testing"
)

// MockDataManager creates a mock DataManager for testing
func NewCommandsMockDataManager() *files.DataManager {
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

// TestSetPower tests setting power
func TestSetPower(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "power",
			offset:  232,
			mask:    65535,
			minimum: 0,
			maximum: 1,
		},
	}

	SetPower("test")
	// Test passes if no panic occurs
}

// TestSetPowerEmptyCommand tests empty power command
func TestSetPowerEmptyCommand(t *testing.T) {
	dm = NewCommandsMockDataManager()

	SetPower("")
	// Should handle gracefully without panic
}

// TestSetParameterAboveMaximum tests parameter exceeding maximum value
func TestSetParameterAboveMaximum(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "setting",
			offset:  15,
			mask:    255,
			minimum: 0,
			maximum: 50,
		},
	}

	SetParameterByTitle("setting", "100")
	// Should fail gracefully as 100 > 50
}

// TestSetParameterBelowMinimum tests parameter below minimum value
func TestSetParameterBelowMinimum(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "setting",
			offset:  15,
			mask:    255,
			minimum: 10,
			maximum: 50,
		},
	}

	SetParameterByTitle("setting", "5")
	// Should fail gracefully as 5 < 10
}

// TestSetParameterInvalidValue tests non-numeric parameter value
func TestSetParameterInvalidValue(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "setting",
			offset:  15,
			mask:    255,
			minimum: 0,
			maximum: 100,
		},
	}

	SetParameterByTitle("setting", "not_a_number")
	// Should fail gracefully on parsing error
}

// TestSetParameterNotFound tests setting a parameter that doesn't exist
func TestSetParameterNotFound(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "existing_param",
			offset:  10,
			mask:    255,
			minimum: 0,
			maximum: 100,
		},
	}

	SetParameterByTitle("nonexistent_param", "50")
	// Should fail gracefully without panicking
}

// TestSetParameterNegativeValue tests negative parameter value
func TestSetParameterNegativeValue(t *testing.T) {
	dm = NewCommandsMockDataManager()
	parameters = []parameter{
		{
			title:   "setting",
			offset:  10,
			mask:    255,
			minimum: 0,
			maximum: 100,
		},
	}

	SetParameterByTitle("setting", "-5")
	// Should fail as -5 < minimum of 0
}

// Test that constants have expected values
func TestCommandsConstantsValidity(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"setPowerOn", setPowerOn, "on"},
		{"setPowerOff", setPowerOff, "off"},
		{"statusManagedGet", statusManagedGet, 232},
		{"statusManagedOn", statusManagedOn, 85},
		{"statusManagedOff", statusManagedOff, 170},
	}

	for _, tt := range tests {
		if tt.value != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.value)
		}
	}
}
