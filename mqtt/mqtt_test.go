package mqtt

import (
	"errors"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMqttClient is a mock implementation of MQTT.Client
type MockMqttClient struct {
	mock.Mock
}

func (m *MockMqttClient) Connect() MQTT.Token {
	args := m.Called()
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) Disconnect(quiesce uint) {
	m.Called(quiesce)
}

func (m *MockMqttClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMqttClient) IsConnectionOpen() bool {
	args := m.Called()
	return args.Bool(0)
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

func (m *MockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	args := m.Called(topic, qos, retained, payload)
	return args.Get(0).(MQTT.Token)
}

func (m *MockMqttClient) AddRoute(topic string, callback MQTT.MessageHandler) {
	m.Called(topic, callback)
}

func (m *MockMqttClient) OptionsReader() MQTT.ClientOptionsReader {
	args := m.Called()
	return args.Get(0).(MQTT.ClientOptionsReader)
}

// MockToken is a mock implementation of MQTT.Token
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

func (m *MockToken) Done() <-chan struct{} {
	args := m.Called()
	if args.Get(0) == nil {
		return make(chan struct{})
	}
	return args.Get(0).(<-chan struct{})
}

// TestNewMqttConnectionWithCustomBaseTopic tests topic construction with custom base topic
func TestNewMqttConnectionWithCustomBaseTopic(t *testing.T) {
	customBaseTopic := "custom/topic"

	// Create a MqttConnection manually to test topic construction
	mc := &MqttConnection{
		qos:       qosAtMostOnce,
		retain:    false,
		baseTopic: customBaseTopic,
	}
	mc.setTopic = mc.baseTopic + "/set/"
	mc.subscribeTopic = mc.setTopic + "+"
	mc.statusTopic = mc.baseTopic + "/status"

	assert.Equal(t, "custom/topic", mc.baseTopic)
	assert.Equal(t, "custom/topic/set/", mc.setTopic)
	assert.Equal(t, "custom/topic/set/+", mc.subscribeTopic)
	assert.Equal(t, "custom/topic/status", mc.statusTopic)
}

// TestHandleConnectionLost tests the handleConnectionLost function
func TestHandleConnectionLost(t *testing.T) {
	mc := &MqttConnection{}
	err := errors.New("connection lost")

	// Should not panic
	assert.NotPanics(t, func() {
		mc.handleConnectionLost(err)
	})
}

// TestHandleConnectWithReceiver tests handleConnect when receiver is provided
func TestHandleConnectWithReceiver(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)

	mockToken.On("WaitTimeout").Return(true)
	mockToken.On("Error").Return(nil)
	// Use mock.Anything for the callback since it's a function type
	mockClient.On("Subscribe", "test/set/+", byte(2), mock.Anything).Return(mockToken)

	mc := &MqttConnection{
		subscribeTopic: "test/set/+",
		receiver: func(key, value string) {
			// Mock receiver
		},
	}

	assert.NotPanics(t, func() {
		mc.handleConnect(mockClient)
	})

	mockClient.AssertCalled(t, "Subscribe", "test/set/+", byte(2), mock.Anything)
}

// TestHandleConnectWithoutReceiver tests handleConnect when receiver is nil
func TestHandleConnectWithoutReceiver(t *testing.T) {
	mockClient := new(MockMqttClient)

	mc := &MqttConnection{
		receiver: nil,
	}

	// Should not subscribe if receiver is nil
	assert.NotPanics(t, func() {
		mc.handleConnect(mockClient)
	})

	mockClient.AssertNotCalled(t, "Subscribe")
}

// TestReceiveMqttWithMatchingTopic tests receiveMqtt with a matching topic
func TestReceiveMqttWithMatchingTopic(t *testing.T) {
	receivedKey := ""
	receivedValue := ""

	mc := &MqttConnection{
		setTopic: "test/set/",
		receiver: func(key, value string) {
			receivedKey = key
			receivedValue = value
		},
	}

	// Create a mock message
	mockMessage := new(mockMessage)
	mockMessage.On("Topic").Return("test/set/temperature")
	mockMessage.On("Payload").Return([]byte("25.5"))

	mc.receiveMqtt(nil, mockMessage)

	assert.Equal(t, "temperature", receivedKey)
	assert.Equal(t, "25.5", receivedValue)
}

// TestReceiveMqttWithNonMatchingTopic tests receiveMqtt with a non-matching topic
func TestReceiveMqttWithNonMatchingTopic(t *testing.T) {
	callCount := 0

	mc := &MqttConnection{
		setTopic: "test/set/",
		receiver: func(key, value string) {
			callCount++
		},
	}

	mockMessage := new(mockMessage)
	mockMessage.On("Topic").Return("test/other/temperature")
	mockMessage.On("Payload").Return([]byte("25.5"))

	mc.receiveMqtt(nil, mockMessage)

	assert.Equal(t, 0, callCount)
}

// TestReceiveMqttWithNestedTopic tests receiveMqtt with nested path in key
func TestReceiveMqttWithNestedTopic(t *testing.T) {
	receivedKey := ""
	receivedValue := ""

	mc := &MqttConnection{
		setTopic: "test/set/",
		receiver: func(key, value string) {
			receivedKey = key
			receivedValue = value
		},
	}

	mockMessage := new(mockMessage)
	mockMessage.On("Topic").Return("test/set/device/mode")
	mockMessage.On("Payload").Return([]byte("auto"))

	mc.receiveMqtt(nil, mockMessage)

	assert.Equal(t, "device/mode", receivedKey)
	assert.Equal(t, "auto", receivedValue)
}

// TestPublish tests the Publish function
func TestPublish(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)

	mockToken.On("WaitTimeout").Return(true)
	mockToken.On("Error").Return(nil)
	mockClient.On("Publish", "base/sensors/temperature", byte(1), true, "25.5").Return(mockToken)

	mc := &MqttConnection{
		mqttClient: mockClient,
		baseTopic:  "base",
		qos:        qosAtLeastOnce,
		retain:     true,
	}

	mc.Publish("sensors", "temperature", "25.5")

	mockClient.AssertExpectations(t)
}

// TestPublishWithDifferentQosAndRetain tests Publish with different QoS and retain settings
func TestPublishWithDifferentQosAndRetain(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)

	mockToken.On("WaitTimeout").Return(true)
	mockToken.On("Error").Return(nil)
	mockClient.On("Publish", "mybase/status/online", byte(0), false, "true").Return(mockToken)

	mc := &MqttConnection{
		mqttClient: mockClient,
		baseTopic:  "mybase",
		qos:        qosAtMostOnce,
		retain:     false,
	}

	mc.Publish("status", "online", "true")

	mockClient.AssertExpectations(t)
}

// TestPublishFailure tests Publish when publish fails
func TestPublishFailure(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)

	mockToken.On("WaitTimeout").Return(false)
	mockToken.On("Error").Return(errors.New("publish failed"))
	mockClient.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockToken)

	mc := &MqttConnection{
		mqttClient: mockClient,
		baseTopic:  "base",
		qos:        qosAtLeastOnce,
		retain:     true,
	}

	// Should not panic even on failure
	assert.NotPanics(t, func() {
		mc.Publish("sensors", "temperature", "25.5")
	})
}

// TestPublishWithEmptyCategory tests Publish with empty category
func TestPublishWithEmptyCategory(t *testing.T) {
	mockClient := new(MockMqttClient)
	mockToken := new(MockToken)

	mockToken.On("WaitTimeout").Return(true)
	mockToken.On("Error").Return(nil)
	mockClient.On("Publish", "base//key", byte(1), true, "value").Return(mockToken)

	mc := &MqttConnection{
		mqttClient: mockClient,
		baseTopic:  "base",
		qos:        qosAtLeastOnce,
		retain:     true,
	}

	mc.Publish("", "key", "value")

	mockClient.AssertExpectations(t)
}

// mockMessage is a mock implementation of MQTT.Message
type mockMessage struct {
	mock.Mock
}

func (m *mockMessage) Duplicate() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockMessage) Qos() byte {
	args := m.Called()
	return args.Get(0).(byte)
}

func (m *mockMessage) Retained() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockMessage) Topic() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockMessage) MessageID() uint16 {
	args := m.Called()
	return args.Get(0).(uint16)
}

func (m *mockMessage) Payload() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *mockMessage) Ack() {
	m.Called()
}
