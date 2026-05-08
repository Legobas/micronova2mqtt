package mqtt

import (
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	connectionTimeout = time.Second * 10
	topicBase         = "micronova2mqtt/"
	topicSet          = topicBase + "set/"
	topicSubscribe    = topicSet + "+"
	topicStatus       = topicBase + "status"
	topicParameters   = topicBase + "parameters"
	qosAtMostOnce     = byte(0)
	qosAtLeastOnce    = byte(1)
	qosExactlyOnce    = byte(2)
	retainMessages    = true
)

type receiveFunc func(key, value string)

type MqttProperties struct {
	Url      string
	User     string
	Password string
	Qos      byte
	Retain   bool
	ClientId string
	Receiver receiveFunc
}

type MqttConnection struct {
	mqttClient MQTT.Client
	qos        byte
	retain     bool
	configPath string
	receiver   receiveFunc
}

func NewMqttConnection(properties MqttProperties) (*MqttConnection, error) {
	mc := &MqttConnection{}
	mc.qos = properties.Qos
	mc.retain = properties.Retain
	mc.receiver = properties.Receiver

	opts := MQTT.NewClientOptions().
		AddBroker(properties.Url).
		SetClientID(properties.ClientId).
		SetCleanSession(true).
		SetBinaryWill(topicStatus, []byte("Offline"), qosAtMostOnce, retainMessages).
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

	token = mc.mqttClient.Publish(topicStatus, qosExactlyOnce, retainMessages, "Online")
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

	// subscribe only if receiver provided
	if mc.receiver != nil {
		log.Info().Str("topic", topicSubscribe).Msg("Subscribing to")

		token := c.Subscribe(topicSubscribe, qosExactlyOnce, mc.receiveMqtt)
		if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", topicSubscribe).Msg("Subscription failed")
		}
	}
}

func (mc MqttConnection) receiveMqtt(client MQTT.Client, msg MQTT.Message) {
	topic := msg.Topic()
	if strings.HasPrefix(topic, topicSet) {
		key := strings.TrimPrefix(topic, topicSet)
		value := string(msg.Payload())
		mc.receiver(key, value)
	}
}

func (mc MqttConnection) Publish(key string, value string) {
	topic := topicParameters + "/" + key
	token := mc.mqttClient.Publish(topic, mc.qos, mc.retain, value)
	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Error().Err(token.Error()).Msgf("Failed to publish to %s", topic)
	}
}
