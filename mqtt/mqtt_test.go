package mqtt

import (
	"errors"
	"strings"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// ---- fakes ----

type fakeToken struct {
	MQTT.Token // embedded so unused methods are satisfied without implementing them
	err        error
	timedOut   bool
}

func (t *fakeToken) Wait() bool                     { return true }
func (t *fakeToken) WaitTimeout(time.Duration) bool { return !t.timedOut }
func (t *fakeToken) Error() error                   { return t.err }

type fakeMessage struct {
	MQTT.Message
	topic   string
	payload []byte
}

func (m fakeMessage) Topic() string   { return m.topic }
func (m fakeMessage) Payload() []byte { return m.payload }

type publishCall struct {
	topic    string
	qos      byte
	retained bool
	payload  interface{}
}

type fakeClient struct {
	MQTT.Client
	connected   bool
	token       *fakeToken
	published   []publishCall
	subscribed  []string
	subQos      []byte
	handlers    map[string]MQTT.MessageHandler
	disconnects []uint
}

func newFakeClient(connected bool) *fakeClient {
	return &fakeClient{
		connected: connected,
		token:     &fakeToken{},
		handlers:  map[string]MQTT.MessageHandler{},
	}
}

func (c *fakeClient) IsConnected() bool { return c.connected }

func (c *fakeClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	c.published = append(c.published, publishCall{topic, qos, retained, payload})
	return c.token
}

func (c *fakeClient) Subscribe(topic string, qos byte, cb MQTT.MessageHandler) MQTT.Token {
	c.subscribed = append(c.subscribed, topic)
	c.subQos = append(c.subQos, qos)
	c.handlers[topic] = cb
	return c.token
}

func (c *fakeClient) Disconnect(quiesce uint) {
	c.disconnects = append(c.disconnects, quiesce)
}

func newTestConn(fc *fakeClient) *MqttConnection {
	return &MqttConnection{
		mqttClient:    fc,
		qos:           QosAtLeastOnce,
		retain:        true,
		baseTopic:     "home/sensors",
		statusTopic:   "home/sensors/status",
		subscriptions: make(map[string]SubscriberFunc),
	}
}

// ---- NewMqttConnection ----

func TestNewMqttConnection_EmptyBaseTopic(t *testing.T) {
	for _, base := range []string{"", "   ", "/"} {
		_, err := NewMqttConnection(MqttProperties{Url: "tcp://127.0.0.1:1", BaseTopic: base})
		if err == nil {
			t.Errorf("BaseTopic %q: expected error, got nil", base)
		}
	}
}

func TestNewMqttConnection_InvalidQos(t *testing.T) {
	_, err := NewMqttConnection(MqttProperties{
		Url:       "tcp://127.0.0.1:1",
		BaseTopic: "x",
		Qos:       QosExactlyOnce + 1,
	})
	if err == nil || !strings.Contains(err.Error(), "invalid QoS") {
		t.Fatalf("expected invalid QoS error, got %v", err)
	}
}

func TestNewMqttConnection_UnreachableBroker(t *testing.T) {
	// Port 1 is almost never open, so the initial connect should fail.
	_, err := NewMqttConnection(MqttProperties{
		Url:       "tcp://127.0.0.1:1",
		BaseTopic: "x",
		ClientId:  "test",
	})
	if err == nil || !strings.Contains(err.Error(), "MQTT connection failed") {
		t.Fatalf("expected connection failure, got %v", err)
	}
}

// ---- waitToken ----

func TestWaitToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := waitToken(&fakeToken{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("token error", func(t *testing.T) {
		want := errors.New("boom")
		if err := waitToken(&fakeToken{err: want}); !errors.Is(err, want) {
			t.Fatalf("expected %v, got %v", want, err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		err := waitToken(&fakeToken{timedOut: true})
		if err == nil || !strings.Contains(err.Error(), "timed out") {
			t.Fatalf("expected timeout error, got %v", err)
		}
	})
}

// ---- Publish ----

func TestPublish_UsesConfiguredQosAndRetain(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)

	mc.Publish("home/sensors/temp", "21.5")

	if len(fc.published) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(fc.published))
	}
	p := fc.published[0]
	if p.topic != "home/sensors/temp" || p.payload != "21.5" {
		t.Errorf("unexpected topic/payload: %q / %v", p.topic, p.payload)
	}
	if p.qos != QosAtLeastOnce || !p.retained {
		t.Errorf("expected qos=%d retained=true, got qos=%d retained=%v", QosAtLeastOnce, p.qos, p.retained)
	}
}

func TestPublish_ErrorDoesNotPanic(t *testing.T) {
	fc := newFakeClient(true)
	fc.token = &fakeToken{err: errors.New("network down")}
	mc := newTestConn(fc)

	mc.Publish("a/b", "x") // should log and return, not panic
	if len(fc.published) != 1 {
		t.Fatalf("expected publish attempt, got %d", len(fc.published))
	}
}

// ---- Subscribe ----

func TestSubscribe_NilHandlerIsIgnored(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)

	mc.Subscribe("a/b", nil)

	if _, ok := mc.subscriptions["a/b"]; ok {
		t.Error("nil handler should not be stored")
	}
	if len(fc.subscribed) != 0 {
		t.Error("nil handler should not trigger a broker subscribe")
	}
}

func TestSubscribe_WhenConnected(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)

	mc.Subscribe("a/b", func(string, string) {})

	if _, ok := mc.subscriptions["a/b"]; !ok {
		t.Error("subscription should be stored")
	}
	if len(fc.subscribed) != 1 || fc.subscribed[0] != "a/b" {
		t.Fatalf("expected subscribe to a/b, got %v", fc.subscribed)
	}
	if fc.subQos[0] != QosAtLeastOnce {
		t.Errorf("expected qos %d, got %d", QosAtLeastOnce, fc.subQos[0])
	}
}

func TestSubscribe_WhenDisconnected_StoresOnly(t *testing.T) {
	fc := newFakeClient(false)
	mc := newTestConn(fc)

	mc.Subscribe("a/b", func(string, string) {})

	if _, ok := mc.subscriptions["a/b"]; !ok {
		t.Error("subscription should be stored for later restore")
	}
	if len(fc.subscribed) != 0 {
		t.Error("should not subscribe on broker while disconnected")
	}
}

func TestSubscribe_HandlerReceivesMessages(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)

	var gotTopic, gotPayload string
	mc.Subscribe("a/b", func(topic, payload string) {
		gotTopic, gotPayload = topic, payload
	})

	cb := fc.handlers["a/b"]
	if cb == nil {
		t.Fatal("expected a message callback to be registered")
	}
	cb(fc, fakeMessage{topic: "a/b", payload: []byte("hello")})

	if gotTopic != "a/b" || gotPayload != "hello" {
		t.Errorf("handler got topic=%q payload=%q", gotTopic, gotPayload)
	}
}

// ---- handleConnect ----

func TestHandleConnect_PublishesOnlineAndRestoresSubscriptions(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)
	mc.subscriptions["a/b"] = func(string, string) {}
	mc.subscriptions["c/d"] = func(string, string) {}

	mc.handleConnect(fc)

	if len(fc.published) == 0 {
		t.Fatal("expected online status publish")
	}
	status := fc.published[0]
	if status.topic != "home/sensors/status" || status.payload != "Online" {
		t.Errorf("unexpected status publish: %q %v", status.topic, status.payload)
	}
	if status.qos != QosAtLeastOnce || !status.retained {
		t.Errorf("status should be retained QoS 1, got qos=%d retained=%v", status.qos, status.retained)
	}

	if len(fc.subscribed) != 2 {
		t.Fatalf("expected 2 restored subscriptions, got %v", fc.subscribed)
	}
}

// ---- Close ----

func TestClose_WhenConnected_PublishesOfflineAndDisconnects(t *testing.T) {
	fc := newFakeClient(true)
	mc := newTestConn(fc)

	mc.Close()

	if len(fc.published) != 1 {
		t.Fatalf("expected offline publish, got %d", len(fc.published))
	}
	p := fc.published[0]
	if p.topic != "home/sensors/status" || p.payload != "Offline" || !p.retained {
		t.Errorf("unexpected offline publish: %+v", p)
	}
	if len(fc.disconnects) != 1 || fc.disconnects[0] != 250 {
		t.Errorf("expected disconnect(250), got %v", fc.disconnects)
	}
}

func TestClose_WhenDisconnected_SkipsPublish(t *testing.T) {
	fc := newFakeClient(false)
	mc := newTestConn(fc)

	mc.Close()

	if len(fc.published) != 0 {
		t.Error("should not publish offline status when disconnected")
	}
	if len(fc.disconnects) != 1 {
		t.Errorf("expected disconnect call, got %v", fc.disconnects)
	}
}

func TestClose_NilReceiverIsSafe(t *testing.T) {
	var mc *MqttConnection
	mc.Close() // must not panic
}

// ---- IsConnected ----

func TestIsConnected(t *testing.T) {
	if (&MqttConnection{}).IsConnected() {
		t.Error("nil client should not be connected")
	}
	if !newTestConn(newFakeClient(true)).IsConnected() {
		t.Error("expected connected")
	}
	if newTestConn(newFakeClient(false)).IsConnected() {
		t.Error("expected disconnected")
	}
}
