package homeassistant

import (
	"encoding/json"
	"fmt"
	"micronova2mqtt/mqtt"
	"testing"
)

type mockQueue struct {
	published  map[string]string
	subscribed map[string]mqtt.SubscriberFunc
}

func newMockQueue() *mockQueue {
	return &mockQueue{
		published:  make(map[string]string),
		subscribed: make(map[string]mqtt.SubscriberFunc),
	}
}

func (m *mockQueue) Publish(topic, value string) {
	m.published[topic] = value
}

func (m *mockQueue) Subscribe(topic string, handler mqtt.SubscriberFunc) {
	m.subscribed[topic] = handler
}

func TestReset_Online(t *testing.T) {
	mq := newMockQueue()

	entity := HAEntityConfig{
		Component:  "sensor",
		UniqueId:   "room_temperature",
		Name:       "Room Temperature", // Set this so your debug log isn't blank
		StateTopic: "micronova/state/temp",
		Device: HADevice{
			Name: "Micronova Stove",
		},
	}

	discovery := HADiscovery{
		DeviceName: "Stove",
		Entities:   []HAEntityConfig{entity},
		Publisher:  mq,
	}

	// Act
	discovery.reset(HAStatusTopic, HAOnline)

	// Assert: Constructing the expected topic precisely matches your formatting template
	expectedTopic := fmt.Sprintf("homeassistant/%s/%s/%s/config", entity.Component, discovery.DeviceName, entity.UniqueId)

	payload, exists := mq.published[expectedTopic]
	if !exists {
		t.Fatalf("Expected configuration to be published to %s, but found nothing", expectedTopic)
	}

	// Verify payload schema
	var parsedConfig HAEntityConfig
	if err := json.Unmarshal([]byte(payload), &parsedConfig); err != nil {
		t.Fatalf("Failed to parse published JSON payload: %v", err)
	}

	if parsedConfig.UniqueId != entity.UniqueId {
		t.Errorf("Expected UniqueId %s, got %s", entity.UniqueId, parsedConfig.UniqueId)
	}

	if parsedConfig.Name != entity.Name {
		t.Errorf("Expected Name %s, got %s", entity.Name, parsedConfig.Name)
	}
}
