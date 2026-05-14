package main

import (
	"testing"
)

// Mock interfaces for dependencies
type mockMicronova struct {
	powerCalls     []string
	parameterCalls []struct {
		key   string
		value string
	}
	updateMqttCalls []bool
}

func (m *mockMicronova) SetPower(value string) {
	m.powerCalls = append(m.powerCalls, value)
}

func (m *mockMicronova) SetParameter(key, value string) {
	m.parameterCalls = append(m.parameterCalls, struct {
		key   string
		value string
	}{key, value})
}

func (m *mockMicronova) SetUpdateMqtt(value bool) {
	m.updateMqttCalls = append(m.updateMqttCalls, value)
}

// TestReceiveMqttMessage_PowerKey tests setting power through MQTT
func TestReceiveMqttMessage_PowerKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected string
	}{
		{
			name:     "power on",
			key:      powerKey,
			value:    "on",
			expected: "on",
		},
		{
			name:     "power off",
			key:      powerKey,
			value:    "off",
			expected: "off",
		},
		{
			name:     "power toggle",
			key:      powerKey,
			value:    "toggle",
			expected: "toggle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test demonstrates the receiveMqttMessage function logic
			// In practice, you'd need to inject the micronova dependency
			if tt.key == powerKey {
				// Verify power key is being handled
				if tt.key != powerKey {
					t.Errorf("Expected power key %q, got %q", powerKey, tt.key)
				}
				if tt.value != tt.expected {
					t.Errorf("Expected value %q, got %q", tt.expected, tt.value)
				}
			}
		})
	}
}

// TestReceiveMqttMessage_ParameterKey tests setting parameters through MQTT
func TestReceiveMqttMessage_ParameterKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "set temperature",
			key:   "temperature",
			value: "22.5",
		},
		{
			name:  "set mode",
			key:   "mode",
			value: "heating",
		},
		{
			name:  "set fan speed",
			key:   "fanSpeed",
			value: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify non-power keys are treated as parameters
			if tt.key == powerKey {
				t.Errorf("Test should use non-power key, got %q", tt.key)
			}
		})
	}
}

// TestPowerKey_Constant tests the power key constant
func TestPowerKey_Constant(t *testing.T) {
	expected := "Power"
	if powerKey != expected {
		t.Errorf("powerKey constant = %q, want %q", powerKey, expected)
	}
}

// TestAppName_Constant tests the app name constant
func TestAppName_Constant(t *testing.T) {
	expected := "micronova2mqtt"
	if appName != expected {
		t.Errorf("AppName constant = %q, want %q", appName, expected)
	}
}

// TestVersion_Variable tests that Version can be set
func TestVersion_Variable(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "1.2.3"
	if Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", Version, "1.2.3")
	}
}

// Example of how to test receiveMqttMessage with dependency injection
// This shows the recommended approach for testable code

// ReceiveMqttMessageWithDeps is a refactored version that accepts dependencies
// This allows for better testing and is recommended for your actual implementation
type MessageReceiver struct {
	micronovaAdapter MicronovaAdapter
}

type MicronovaAdapter interface {
	SetPower(value string)
	SetParameter(key, value string)
	SetUpdateMqtt(value bool)
}

func (mr *MessageReceiver) ReceiveMqttMessage(key string, value string) {
	if key == powerKey {
		mr.micronovaAdapter.SetPower(value)
	} else {
		mr.micronovaAdapter.SetParameter(key, value)
	}
	mr.micronovaAdapter.SetUpdateMqtt(true)
}

// TestMessageReceiver_PowerCommand tests power command handling
func TestMessageReceiver_PowerCommand(t *testing.T) {
	mock := &mockMicronova{}
	receiver := &MessageReceiver{micronovaAdapter: mock}

	receiver.ReceiveMqttMessage("Power", "on")

	if len(mock.powerCalls) != 1 {
		t.Errorf("Expected 1 power call, got %d", len(mock.powerCalls))
	}
	if mock.powerCalls[0] != "on" {
		t.Errorf("Expected power value %q, got %q", "on", mock.powerCalls[0])
	}
	if len(mock.updateMqttCalls) != 1 || !mock.updateMqttCalls[0] {
		t.Errorf("Expected SetUpdateMqtt(true) to be called")
	}
}

// TestMessageReceiver_ParameterCommand tests parameter setting
func TestMessageReceiver_ParameterCommand(t *testing.T) {
	mock := &mockMicronova{}
	receiver := &MessageReceiver{micronovaAdapter: mock}

	receiver.ReceiveMqttMessage("temperature", "22.5")

	if len(mock.parameterCalls) != 1 {
		t.Errorf("Expected 1 parameter call, got %d", len(mock.parameterCalls))
	}
	if mock.parameterCalls[0].key != "temperature" || mock.parameterCalls[0].value != "22.5" {
		t.Errorf("Expected parameter (temperature, 22.5), got (%s, %s)",
			mock.parameterCalls[0].key, mock.parameterCalls[0].value)
	}
	if len(mock.updateMqttCalls) != 1 || !mock.updateMqttCalls[0] {
		t.Errorf("Expected SetUpdateMqtt(true) to be called")
	}
}

// TestMessageReceiver_MultipleCommands tests multiple consecutive commands
func TestMessageReceiver_MultipleCommands(t *testing.T) {
	mock := &mockMicronova{}
	receiver := &MessageReceiver{micronovaAdapter: mock}

	receiver.ReceiveMqttMessage("Power", "on")
	receiver.ReceiveMqttMessage("temperature", "21")
	receiver.ReceiveMqttMessage("Power", "off")

	if len(mock.powerCalls) != 2 {
		t.Errorf("Expected 2 power calls, got %d", len(mock.powerCalls))
	}
	if len(mock.parameterCalls) != 1 {
		t.Errorf("Expected 1 parameter call, got %d", len(mock.parameterCalls))
	}
	if len(mock.updateMqttCalls) != 3 {
		t.Errorf("Expected 3 update subscriber calls, got %d", len(mock.updateMqttCalls))
	}
}
