package mqtt

import (
	"errors"
	"sync"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// --- Mocks ---

type mockToken struct {
	err error
}

func (m *mockToken) Wait() bool                       { return true }
func (m *mockToken) WaitTimeout(d time.Duration) bool { return true }
func (m *mockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
func (m *mockToken) Error() error { return m.err }

type publishedMsg struct {
	topic    string
	qos      byte
	retained bool
	payload  interface{}
}

type mockClient struct {
	mu           sync.Mutex
	connected    bool
	published    []publishedMsg
	subscribed   map[string]MQTT.MessageHandler
	publishErr   error
	subscribeErr error
}

func newMockClient() *mockClient {
	return &mockClient{
		connected:  true,
		subscribed: make(map[string]MQTT.MessageHandler),
	}
}

func (m *mockClient) IsConnected() bool      { return m.connected }
func (m *mockClient) IsConnectionOpen() bool { return m.connected }
func (m *mockClient) Connect() MQTT.Token    { return &mockToken{} }
func (m *mockClient) Disconnect(quiesce uint) {
	m.connected = false
}

func (m *mockClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, publishedMsg{
		topic:    topic,
		qos:      qos,
		retained: retained,
		payload:  payload,
	})
	return &mockToken{err: m.publishErr}
}

func (m *mockClient) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) MQTT.Token {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribed[topic] = callback
	return &mockToken{err: m.subscribeErr}
}

func (m *mockClient) SubscribeMultiple(filters map[string]byte, callback MQTT.MessageHandler) MQTT.Token {
	return &mockToken{}
}
func (m *mockClient) Unsubscribe(topics ...string) MQTT.Token             { return &mockToken{} }
func (m *mockClient) AddRoute(topic string, callback MQTT.MessageHandler) {}
func (m *mockClient) OptionsReader() MQTT.ClientOptionsReader             { return MQTT.ClientOptionsReader{} }

// mockMessage implements MQTT.Message to simulate incoming broker messages
type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool   { return false }
func (m *mockMessage) Qos() byte         { return 0 }
func (m *mockMessage) Retained() bool    { return false }
func (m *mockMessage) Topic() string     { return m.topic }
func (m *mockMessage) MessageID() uint16 { return 1 }
func (m *mockMessage) Payload() []byte   { return m.payload }
func (m *mockMessage) Ack()              {}

// --- Unit Tests ---

func setupTestConnection(client *mockClient) *MqttConnection {
	return &MqttConnection{
		mqttClient:    client,
		qos:           QosAtLeastOnce,
		retain:        false,
		baseTopic:     "micronova",
		statusTopic:   "micronova/status",
		subscriptions: make(map[string]SubscriberFunc),
	}
}

func TestPublish(t *testing.T) {
	client := newMockClient()
	mc := setupTestConnection(client)

	mc.Publish("micronova/command", "ON")

	if len(client.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(client.published))
	}

	msg := client.published[0]
	if msg.topic != "micronova/command" {
		t.Errorf("expected topic 'micronova/command', got '%s'", msg.topic)
	}
	if msg.payload != "ON" {
		t.Errorf("expected payload 'ON', got '%v'", msg.payload)
	}
	if msg.qos != QosAtLeastOnce {
		t.Errorf("expected QoS %d, got %d", QosAtLeastOnce, msg.qos)
	}
}

func TestSubscribeAndReceiveMessage(t *testing.T) {
	client := newMockClient()
	mc := setupTestConnection(client)

	receivedTopic := ""
	receivedPayload := ""

	mc.Subscribe("micronova/set", func(topic string, payload string) {
		receivedTopic = topic
		receivedPayload = payload
	})

	// Verify subscription was registered on the client and saved in internal map
	handler, exists := client.subscribed["micronova/set"]
	if !exists {
		t.Fatal("expected subscription to be registered on MQTT client")
	}

	// Simulate an incoming message
	handler(client, &mockMessage{topic: "micronova/set", payload: []byte("22.5")})

	if receivedTopic != "micronova/set" || receivedPayload != "22.5" {
		t.Errorf("handler received (%s, %s), expected ('micronova/set', '22.5')", receivedTopic, receivedPayload)
	}
}

func TestSubscribeNilHandlerIgnored(t *testing.T) {
	client := newMockClient()
	mc := setupTestConnection(client)

	mc.Subscribe("micronova/nil", nil)

	if len(mc.subscriptions) != 0 {
		t.Errorf("expected 0 stored subscriptions for nil handler, got %d", len(mc.subscriptions))
	}
	if len(client.subscribed) != 0 {
		t.Errorf("expected 0 client subscriptions for nil handler, got %d", len(client.subscribed))
	}
}

func TestReconnectRestoresStatusAndSubscriptions(t *testing.T) {
	client := newMockClient()
	mc := setupTestConnection(client)

	// 1. Register subscriptions before disconnect
	called := false
	mc.Subscribe("micronova/power", func(topic string, payload string) {
		called = true
	})

	// 2. Simulate connection drop and broker wiping clean session state
	mc.handleConnectionLost(errors.New("EOF"))
	client.subscribed = make(map[string]MQTT.MessageHandler)
	client.published = nil

	// 3. Simulate Paho triggering OnConnectHandler upon reconnect
	mc.handleConnect(client)

	// Verify "Online" status was republished
	if len(client.published) != 1 {
		t.Fatalf("expected 'Online' status to be republished on reconnect, got %d messages", len(client.published))
	}
	statusMsg := client.published[0]
	if statusMsg.topic != "micronova/status" || statusMsg.payload != "Online" || !statusMsg.retained {
		t.Errorf("unexpected status message on reconnect: %+v", statusMsg)
	}

	// Verify subscription was restored
	restoredHandler, exists := client.subscribed["micronova/power"]
	if !exists {
		t.Fatal("expected 'micronova/power' subscription to be restored on reconnect")
	}

	// Verify restored callback works
	restoredHandler(client, &mockMessage{topic: "micronova/power", payload: []byte("1")})
	if !called {
		t.Error("expected restored subscription handler to execute")
	}
}
