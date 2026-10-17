package mqtt

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	connectionTimeout = time.Second * 10
	qosAtMostOnce     = byte(0)
	qosAtLeastOnce    = byte(1)
	qosExactlyOnce    = byte(2)
	retainMessage     = true
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
	mqttClient  MQTT.Client
	qos         byte
	retain      bool
	baseTopic   string
	statusTopic string
	configPath  string
}

func NewMqttConnection(properties MqttProperties) (*MqttConnection, error) {
	mc := &MqttConnection{}
	mc.qos = properties.Qos
	mc.retain = properties.Retain
	mc.baseTopic = properties.BaseTopic
	mc.statusTopic = mc.baseTopic + "/status"

	opts := MQTT.NewClientOptions().
		AddBroker(properties.Url).
		SetClientID(properties.ClientId).
		SetCleanSession(true).
		SetBinaryWill(mc.statusTopic, []byte("Offline"), qosAtMostOnce, retainMessage).
		SetAutoReconnect(true).
		SetConnectionLostHandler(func(c MQTT.Client, err error) {
			mc.handleConnectionLost(err)
		}).SetOnConnectHandler(func(c MQTT.Client) {
		mc.handleConnect(c)
	})
	if properties.User != "" && properties.Password != "" {
		opts.SetUsername(properties.User)
		opts.SetPassword(properties.Password)
	}

	mc.mqttClient = MQTT.NewClient(opts)
	token := mc.mqttClient.Connect()
	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("MQTT connection failed")
	}

	token = mc.mqttClient.Publish(mc.statusTopic, qosExactlyOnce, retainMessage, "Online")
	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Error().Err(token.Error()).Msg("Failed to publish LWT status")
	}

	return mc, nil
}

func (mc MqttConnection) handleConnectionLost(err error) {
	log.Error().Err(err).Msg("MQTT connection lost, attempting reconnect")
}

func (mc MqttConnection) handleConnect(c MQTT.Client) {
	log.Info().Msg("MQTT client connected")
}

func (mc MqttConnection) Publish(topic string, value string) {
	token := mc.mqttClient.Publish(topic, mc.qos, mc.retain, value)
	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Error().Err(token.Error()).Msgf("Failed to publish raw to %s", topic)
	}
}

func (mc *MqttConnection) Subscribe(topic string, handler SubscriberFunc) {
	if handler == nil {
		log.Warn().Str("topic", topic).Msg("Subscription error: no handler provided")
		return
	}

	log.Info().Str("topic", topic).Msg("Subscribing to")

	mqttCallback := func(client MQTT.Client, message MQTT.Message) {
		handler(message.Topic(), string(message.Payload()))
	}

	token := mc.mqttClient.Subscribe(topic, qosExactlyOnce, mqttCallback)

	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Error().Err(token.Error()).Str("topic", topic).Msg("Subscription failed")
	}
}
