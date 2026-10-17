package homeassistant

import (
	"encoding/json"
	"fmt"
	"micronova2mqtt/mqtt"

	"github.com/rs/zerolog/log"
)

const (
	HABaseTopic   = "homeassistant"
	HAStatusTopic = "homeassistant/status"
	HAOnline      = "online"
)

type Publisher interface {
	Publish(topic, value string)
	Subscribe(topic string, handler mqtt.SubscriberFunc)
}

type HADevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
}

type HAEntityConfig struct {
	Component         string   `json:"-"` // Internal helper (sensor, switch, etc.)
	Name              string   `json:"name"`
	DeviceClass       string   `json:"device_class,omitempty"`
	Options           []string `json:"options,omitempty"`
	ObjectId          string   `json:"object_id,omitempty"`
	UniqueId          string   `json:"unique_id"`
	Icon              string   `json:"icon,omitempty"`
	StateTopic        string   `json:"state_topic"`
	StateOn           string   `json:"state_on,omitempty"`
	StateOff          string   `json:"state_off,omitempty"`
	CommandTopic      string   `json:"command_topic,omitempty"`
	CommandTemplate   string   `json:"command_template,omitempty"`
	PayloadOn         string   `json:"payload_on,omitempty"`
	PayloadOff        string   `json:"payload_off,omitempty"`
	ValueTemplate     string   `json:"value_template,omitempty"`
	UnitOfMeasurement string   `json:"unit_of_measurement,omitempty"`
	Min               int      `json:"min,omitempty"`
	Max               int      `json:"max,omitempty"`
	Step              int      `json:"step,omitempty"`
	Device            HADevice `json:"device"`
}

type HADiscovery struct {
	DeviceName string
	Entities   []HAEntityConfig
	Publisher  Publisher
}

func (discovery HADiscovery) reset(topic string, payload string) {
	log.Warn().Msgf("Home Assistant Status: %s", payload)

	if payload == HAOnline {
		discovery.discover()
	}
}

func (discovery HADiscovery) Setup() {
	discovery.Publisher.Subscribe(HAStatusTopic, discovery.reset)
	discovery.discover()
}

func (discovery HADiscovery) discover() {
	log.Info().Msgf("Registering %d entities with Home Assistant MQTT Auto Discovery...", len(discovery.Entities))

	for _, entity := range discovery.Entities {
		// HA Discovery topic format: homeassistant/<component>/<device>/<parameter>/config
		discoveryTopic := fmt.Sprintf("homeassistant/%s/%s/%s/config", entity.Component, discovery.DeviceName, entity.UniqueId)

		payload, err := json.Marshal(entity)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to marshal HA discovery payload for %s", entity.Name)
			continue
		}

		discovery.Publisher.Publish(discoveryTopic, string(payload))

		log.Debug().Msgf("Dispatched HA discovery for: %s", entity.Name)
	}
}
