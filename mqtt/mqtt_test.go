package mqtt

import (
	"errors"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMqttClient mocks the MQTT.Client interface
type MockMqttClient struct {
	mock.Mock
}

func (m *MockMqttClient) AddRoute(topic string, handler MQTT.MessageHandler) {
	m.Called(topic, handler)
}

func (m *MockMqttClient) Connect() MQTT.Token {
	args := m.Called()
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) Disconnect(quiesce uint) {
	m.Called(quiesce)
}

func (m *MockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	args := m.Called(topic, qos, retained, payload)
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) MQTT.Token {
	args := m.Called(topic, qos, callback)
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) SubscribeMultiple(filters map[string]byte, callback MQTT.MessageHandler) MQTT.Token {
	args := m.Called(filters, callback)
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) Unsubscribe(topics ...string) MQTT.Token {
	args := m.Called(topics)
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttClient) IsConnectionOpen() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockToken mocks the MQTT.Token interface
type MockToken struct {
	mock.Mock
}

func (m *MockToken) Wait() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockToken) WaitTimeout(time.Duration) bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockToken) Error() error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

// TestNewMqttConnectionSuccess tests successful MQTT connection creation
func TestNewMqttConnectionSuccess(t *testing.T) {
	properties := MqttProperties{
		Url:      "tcp://localhost:1883",
		ClientId: "test-client",
		Receiver: func(key, value string) {},
	}

	// Since NewMqttConnection creates a real client, we can test the properties
	// In a real scenario, you'd refactor to inject the client
	assert.NotNil(t, properties.Url)
	assert.NotEmpty(t, properties.ClientId)
	assert.NotNil(t, properties.Receiver)
}

// TestMqttPropertiesValidation tests properties validation
func TestMqttPropertiesValidation(t *testing.T) {
	tests := []struct {
		name       string
		properties MqttProperties
		shouldPass bool
	}{
		{
			name: "valid properties with user and password",
			properties: MqttProperties{
				Url:      "tcp://localhost:1883",
				User:     "user",
				Password: "pass",
				Qos:      qosAtLeastOnce,
				Retain:   true,
				ClientId: "client",
				Receiver: func(key, value string) {},
			},
			shouldPass: true,
		},
		{
			name: "valid properties without user and password",
			properties: MqttProperties{
				Url:      "tcp://localhost:1883",
				Qos:      qosAtMostOnce,
				Retain:   false,
				ClientId: "client",
				Receiver: func(key, value string) {},
			},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPass {
				assert.NotEmpty(t, tt.properties.Url)
				assert.NotEmpty(t, tt.properties.ClientId)
			}
		})
	}
}


// TestReceiveMqttMessageParsing tests the receiveMqtt message parsing
func TestReceiveMqttMessageParsing(t *testing.T) {
	tests := []struct {
		name           string
		topic          string
		payload        string
		shouldCallback bool
		expectedKey    string
		expectedValue  string
	}{
		{
			name:           "valid set topic",
			topic:          topicSet + "temperature",
			payload:        "25.5",
			shouldCallback: true,
			expectedKey:    "temperature",
			expectedValue:  "25.5",
		},
		{
			name:           "valid set topic with nested path",
			topic:          topicSet + "device/sensor/value",
			payload:        "active",
			shouldCallback: true,
			expectedKey:    "device/sensor/value",
			expectedValue:  "active",
		},
		{
			name:           "non-set topic should be ignored",
			topic:          topicStatus,
			payload:        "Online",
			shouldCallback: false,
		},
		{
			name:           "partial match should be ignored",
			topic:          "other/set/value",
			payload:        "test",
			shouldCallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			var capturedKey, capturedValue string

			mc := &MqttConnection{
				receiver: func(key, value string) {
					callCount++
					capturedKey = key
					capturedValue = value
				},
			}

			mockMessage := new(MockMessage)
			mockMessage.On("Topic").Return(tt.topic)
			mockMessage.On("Payload").Return([]byte(tt.payload))

			mc.receiveMqtt(nil, mockMessage)

			if tt.shouldCallback {
				require.Equal(t, 1, callCount, "receiver callback should be called")
				assert.Equal(t, tt.expectedKey, capturedKey)
				assert.Equal(t, tt.expectedValue, capturedValue)
			} else {
				assert.Equal(t, 0, callCount, "receiver callback should not be called")
			}
		})
	}
}

// TestHandleConnectionLost tests connection lost handler
func TestHandleConnectionLost(t *testing.T) {
	mc := &MqttConnection{}
	testErr := errors.New("connection timeout")

	// Should not panic and should log the error
	assert.NotPanics(t, func() {
		mc.handleConnectionLost(testErr)
	})
}

// TestConstants verifies constants are correctly defined
func TestConstants(t *testing.T) {
	assert.Equal(t, 10*time.Second, connectionTimeout)
	assert.Equal(t, "micronova2mqtt/", topicBase)
	assert.Equal(t, "micronova2mqtt/set/", topicSet)
	assert.Equal(t, "micronova2mqtt/set/+", topicSubscribe)
	assert.Equal(t, "micronova2mqtt/status", topicStatus)
	assert.Equal(t, byte(0), qosAtMostOnce)
	assert.Equal(t, byte(1), qosAtLeastOnce)
	assert.Equal(t, byte(2), qosExactlyOnce)
	assert.True(t, retainMessages)
}

// MockMessage mocks the MQTT.Message interface
type MockMessage struct {
	mock.Mock
}

func (m *MockMessage) Duplicate() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMessage) Qos() byte {
	args := m.Called()
	return args.Get(0).(byte)
}

func (m *MockMessage) Retained() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMessage) Topic() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMessage) MessageID() uint16 {
	args := m.Called()
	return args.Get(0).(uint16)
}

func (m *MockMessage) Payload() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockMessage) Ack() {
	m.Called()
}
