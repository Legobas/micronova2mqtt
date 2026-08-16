package mqtt

import (
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	connectionTimeout = time.Second * 10
	topicBase         = "micronova2mqtt"
	qosAtMostOnce     = byte(0)
	qosAtLeastOnce    = byte(1)
	qosExactlyOnce    = byte(2)
	retainMessage     = true
)

type receiveFunc func(key, value string)

type MqttProperties struct {
	Url       string
	User      string
	Password  string
	Qos       byte
	Retain    bool
	BaseTopic string
	ClientId  string
	Receiver  receiveFunc
}

type MqttConnection struct {
	mqttClient     MQTT.Client
	qos            byte
	retain         bool
	baseTopic      string
	setTopic       string
	subscribeTopic string
	statusTopic    string
	configPath     string
	receiver       receiveFunc
}

func NewMqttConnection(properties MqttProperties) (*MqttConnection, error) {
	mc := &MqttConnection{}
	mc.qos = properties.Qos
	mc.retain = properties.Retain
	if len(properties.BaseTopic) == 0 {
		mc.baseTopic = topicBase
	} else {
		mc.baseTopic = properties.BaseTopic
	}
	mc.setTopic = mc.baseTopic + "/set/"
	mc.subscribeTopic = mc.setTopic + "+"
	mc.statusTopic = mc.baseTopic + "/status"
	mc.receiver = properties.Receiver

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

	// subscribe only if receiver provided
	if mc.receiver != nil {
		log.Info().Str("topic", mc.subscribeTopic).Msg("Subscribing to")

		token := c.Subscribe(mc.subscribeTopic, qosExactlyOnce, mc.receiveMqtt)
		if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", mc.subscribeTopic).Msg("Subscription failed")
		}
	}
}

func (mc MqttConnection) receiveMqtt(client MQTT.Client, msg MQTT.Message) {
	topic := msg.Topic()
	if strings.HasPrefix(topic, mc.setTopic) {
		key := strings.TrimPrefix(topic, mc.setTopic)
		value := string(msg.Payload())
		mc.receiver(key, value)
	}
}

func (mc MqttConnection) Publish(category, key, value string) {
	var topic string
	topic = mc.baseTopic + "/" + category + "/" + key
	token := mc.mqttClient.Publish(topic, mc.qos, mc.retain, value)
	if !token.WaitTimeout(connectionTimeout) || token.Error() != nil {
		log.Error().Err(token.Error()).Msgf("Failed to publish to %s", topic)
	}
}
