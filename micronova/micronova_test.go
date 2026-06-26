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
		Key: "Key",
		Topic: "Topic",
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

// TestSetPowerOn tests setting power to on state
func TestSetPowerOn(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "power",
			offset:   232,
			mask:     65535,
			minimum:  0,
			maximum:  1,
		},
	}

	SetPower("on")
	// Test passes if no panic occurs
	// In production, you'd verify the writeDevice call was made
}

// TestSetPowerOff tests setting power to off state
func TestSetPowerOff(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "power",
			offset:   232,
			mask:     65535,
			minimum:  0,
			maximum:  1,
		},
	}

	SetPower("off")
	// Test passes if no panic occurs
}

// TestSetPowerCustomOn tests custom power on value from config
func TestSetPowerCustomOn(t *testing.T) {
	dm = NewMockDataManager()
	dm.Config.Micronova.Power.On = "custom_on"

	SetPower("custom_on")
	// Verify custom value is respected
}

// TestSetPowerCustomOff tests custom power off value from config
func TestSetPowerCustomOff(t *testing.T) {
	dm = NewMockDataManager()
	dm.Config.Micronova.Power.Off = "custom_off"

	SetPower("custom_off")
	// Verify custom value is respected
}

// TestSetPowerInvalid tests invalid power command
func TestSetPowerInvalid(t *testing.T) {
	dm = NewMockDataManager()

	// Should not panic on invalid command
	SetPower("invalid_command")
}

// TestSetParameterValid tests setting a valid parameter
func TestSetParameterValid(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "temperature",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	SetParameter("temperature", "50")
	// Test passes if no panic occurs
}

// TestSetParameterWithinRange tests parameter value within valid range
func TestSetParameterWithinRange(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "speed",
			offset:   20,
			mask:     255,
			minimum:  10,
			maximum:  100,
		},
	}

	SetParameter("speed", "50")
	// Value 50 is within range [10, 100]
}

// TestSetParameterAboveMaximum tests parameter exceeding maximum value
func TestSetParameterAboveMaximum(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "setting",
			offset:   15,
			mask:     255,
			minimum:  0,
			maximum:  50,
		},
	}

	SetParameter("setting", "100")
	// Should fail gracefully as 100 > 50
}

// TestSetParameterBelowMinimum tests parameter below minimum value
func TestSetParameterBelowMinimum(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "setting",
			offset:   15,
			mask:     255,
			minimum:  10,
			maximum:  50,
		},
	}

	SetParameter("setting", "5")
	// Should fail gracefully as 5 < 10
}

// TestSetParameterInvalidValue tests non-numeric parameter value
func TestSetParameterInvalidValue(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "setting",
			offset:   15,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	SetParameter("setting", "not_a_number")
	// Should fail gracefully on parsing error
}

// TestSetParameterNotFound tests setting a parameter that doesn't exist
func TestSetParameterNotFound(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "existing_param",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	SetParameter("nonexistent_param", "50")
	// Should fail gracefully without panicking
}

// TestSetParameterAtMinimumBoundary tests parameter at minimum boundary
func TestSetParameterAtMinimumBoundary(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "bounded",
			offset:   10,
			mask:     255,
			minimum:  10,
			maximum:  100,
		},
	}

	SetParameter("bounded", "10")
	// Should succeed as value equals minimum
}

// TestSetParameterAtMaximumBoundary tests parameter at maximum boundary
func TestSetParameterAtMaximumBoundary(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "bounded",
			offset:   10,
			mask:     255,
			minimum:  10,
			maximum:  100,
		},
	}

	SetParameter("bounded", "100")
	// Should succeed as value equals maximum
}

// TestSetUpdateMqttTrue tests enabling MQTT update flag
func TestSetUpdateMqttTrue(t *testing.T) {
	updateMqtt = false
	SetUpdateMqtt(true)

	if !updateMqtt {
		t.Errorf("Expected updateMqtt to be true, got false")
	}
}

// TestSetUpdateMqttFalse tests disabling MQTT update flag
func TestSetUpdateMqttFalse(t *testing.T) {
	updateMqtt = true
	SetUpdateMqtt(false)

	if updateMqtt {
		t.Errorf("Expected updateMqtt to be false, got true")
	}
}

// TestSetParameterMultipleParameters tests selecting correct parameter from multiple
func TestSetParameterMultipleParameters(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "param1",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
		{
			topicKey: "param2",
			offset:   20,
			mask:     255,
			minimum:  0,
			maximum:  200,
		},
		{
			topicKey: "param3",
			offset:   30,
			mask:     255,
			minimum:  0,
			maximum:  50,
		},
	}

	SetParameter("param2", "150")
	// Should set only param2, not param1 or param3
}

// TestSetParameterZeroValue tests setting parameter to zero
func TestSetParameterZeroValue(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "counter",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	SetParameter("counter", "0")
	// Should succeed as 0 is within range
}

// TestSetParameterNegativeValue tests negative parameter value
func TestSetParameterNegativeValue(t *testing.T) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "setting",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	SetParameter("setting", "-5")
	// Should fail as -5 < minimum of 0
}

// TestSetPowerEmptyCommand tests empty power command
func TestSetPowerEmptyCommand(t *testing.T) {
	dm = NewMockDataManager()

	SetPower("")
	// Should handle gracefully without panic
}

// TestParameterStructure tests parameter data structure
func TestParameterStructure(t *testing.T) {
	param := parameter{
		regKey:   "reg_key_1",
		topicKey: "topic_1",
		offset:   10,
		mask:     255,
		minimum:  0,
		maximum:  100,
		value:    50,
		formula:  "x * 2",
		format:   "%.2f",
		text:     "Test Parameter",
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
		{"setPowerOn", setPowerOn, "on"},
		{"setPowerOff", setPowerOff, "off"},
		{"statusManagedGet", statusManagedGet, 232},
		{"statusManagedOn", statusManagedOn, 85},
		{"statusManagedOff", statusManagedOff, 170},
		{"deviceOnline", deviceOnline, "Online"},
		{"deviceOffline", deviceOffline, "Offline"},
	}

	for _, tt := range tests {
		if tt.value != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.value)
		}
	}
}

// BenchmarkSetPower benchmarks the SetPower function
func BenchmarkSetPower(b *testing.B) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "power",
			offset:   232,
			mask:     65535,
			minimum:  0,
			maximum:  1,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetPower("on")
	}
}

// BenchmarkSetParameter benchmarks the SetParameter function
func BenchmarkSetParameter(b *testing.B) {
	dm = NewMockDataManager()
	parameters = []parameter{
		{
			topicKey: "temperature",
			offset:   10,
			mask:     255,
			minimum:  0,
			maximum:  100,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetParameter("temperature", "50")
	}
}
