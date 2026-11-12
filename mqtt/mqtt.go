package mqtt

import (
	"fmt"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	connectionTimeout    = time.Second * 10
	maxReconnectInterval = time.Second * 30
	QosAtMostOnce        = byte(0)
	QosAtLeastOnce       = byte(1)
	QosExactlyOnce       = byte(2)
	retainMessage        = true
)

type SubscriberFunc func(topic string, payload string)

type MqttProperties struct {
	Url       string
	User      string
	Password  string
	Qos       byte
	Retain    bool
	BaseTopic string
	ClientId  string
}

type MqttConnection struct {
	mqttClient    MQTT.Client
	qos           byte
	retain        bool
	baseTopic     string
	statusTopic   string
	subscriptions map[string]SubscriberFunc
	mu            sync.RWMutex
}

func NewMqttConnection(props MqttProperties) (*MqttConnection, error) {
	base := strings.TrimRight(props.BaseTopic, "/")
	if strings.TrimSpace(base) == "" {
		return nil, fmt.Errorf("MQTT BaseTopic must not be empty")
	}
	if props.Qos > QosExactlyOnce {
		return nil, fmt.Errorf("invalid QoS (%d), must be 0, 1 or 2", props.Qos)
	}

	mc := &MqttConnection{
		qos:           props.Qos,
		retain:        props.Retain,
		baseTopic:     base,
		statusTopic:   base + "/status",
		subscriptions: make(map[string]SubscriberFunc),
	}

	opts := MQTT.NewClientOptions().
		AddBroker(props.Url).
		SetClientID(props.ClientId).
		SetCleanSession(true).
		SetOrderMatters(false).
		SetConnectTimeout(connectionTimeout).
		SetBinaryWill(mc.statusTopic, []byte("Offline"), QosAtLeastOnce, retainMessage).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(maxReconnectInterval).
		SetConnectionLostHandler(func(c MQTT.Client, err error) {
			mc.handleConnectionLost(err)
		}).
		SetOnConnectHandler(func(c MQTT.Client) {
			go mc.handleConnect(c)
		})

	if props.User != "" {
		opts.SetUsername(props.User)
		if props.Password != "" {
			opts.SetPassword(props.Password)
		}
	}

	mc.mqttClient = MQTT.NewClient(opts)
	if err := waitToken(mc.mqttClient.Connect()); err != nil {
		mc.mqttClient.Disconnect(0)
		return nil, fmt.Errorf("MQTT connection failed: %w", err)
	}
	return mc, nil
}

func (mc *MqttConnection) handleConnectionLost(err error) {
	log.Error().Err(err).Msg("MQTT connection lost, attempting reconnect")
}

func (mc *MqttConnection) handleConnect(c MQTT.Client) {
	log.Info().Msg("MQTT client connected")

	token := c.Publish(mc.statusTopic, QosAtLeastOnce, retainMessage, "Online")
	if err := waitToken(token); err != nil {
		log.Error().Err(err).Msg("failed to publish LWT online status")
	}

	mc.mu.RLock()
	subs := make(map[string]SubscriberFunc, len(mc.subscriptions))
	for topic, handler := range mc.subscriptions {
		subs[topic] = handler
	}
	mc.mu.RUnlock()

	for topic, handler := range subs {
		mc.subscribeTopic(c, topic, handler)
	}
}

func (mc *MqttConnection) Publish(topic string, value string) {
	token := mc.mqttClient.Publish(topic, mc.qos, mc.retain, value)
	if err := waitToken(token); err != nil {
		log.Error().Err(err).Msgf("failed to publish to %s", topic)
	}
}

func (mc *MqttConnection) Subscribe(topic string, handler SubscriberFunc) {
	if handler == nil {
		log.Warn().Str("topic", topic).Msg("subscription error: no handler provided")
		return
	}

	// Save subscription so it can be restored after a reconnect
	mc.mu.Lock()
	mc.subscriptions[topic] = handler
	mc.mu.Unlock()

	if mc.mqttClient.IsConnected() {
		mc.subscribeTopic(mc.mqttClient, topic, handler)
	}
}

func (mc *MqttConnection) subscribeTopic(client MQTT.Client, topic string, handler SubscriberFunc) {
	log.Info().Str("topic", topic).Msg("subscribing to topic")

	mqttCallback := func(client MQTT.Client, message MQTT.Message) {
		handler(message.Topic(), string(message.Payload()))
	}

	token := client.Subscribe(topic, mc.qos, mqttCallback)
	if err := waitToken(token); err != nil {
		log.Error().Err(err).Msgf("subscribe to %s failed", topic)
	}
}

func (mc *MqttConnection) Close() {
	if mc.mqttClient.IsConnected() {
		token := mc.mqttClient.Publish(mc.statusTopic, QosAtLeastOnce, retainMessage, "Offline")
		if err := waitToken(token); err != nil {
			log.Error().Err(err).Msg("failed to publish offline status")
		}
	}
	mc.mqttClient.Disconnect(250)
}

func waitToken(token MQTT.Token) error {
	if !token.WaitTimeout(connectionTimeout) {
		return fmt.Errorf("timed out after %s", connectionTimeout)
	}
	return token.Error()
}
