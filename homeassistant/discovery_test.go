package homeassistant

import (
	"encoding/json"
	"micronova2mqtt/mqtt"
	"strings"
	"testing"
)

// MockQueue is een mock-implementatie van de Queue interface
type MockQueue struct {
	Published  map[string]string
	Subscribed map[string]mqtt.SubscriberFunc
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		Published:  make(map[string]string),
		Subscribed: make(map[string]mqtt.SubscriberFunc),
	}
}

func (m *MockQueue) Publish(topic, value string) {
	m.Published[topic] = value
}

func (m *MockQueue) Subscribe(topic string, handler mqtt.SubscriberFunc) {
	m.Subscribed[topic] = handler
}

// TestSetup controleert of de Setup methode correct registreert op het status topic
func TestHADiscovery_Setup(t *testing.T) {
	mockQueue := NewMockQueue()
	discovery := HADiscovery{
		DeviceName: "Test Device",
		Publisher:  mockQueue,
	}

	discovery.Setup()

	// Controleer of er is geabonneerd op het juiste Home Assistant status topic
	handler, exists := mockQueue.Subscribed[HAStatusTopic]
	if !exists {
		t.Fatalf("Verwacht subcriptie op topic %s, maar deze is niet gevonden", HAStatusTopic)
	}
	if handler == nil {
		t.Error("Subscriber handler mag niet nil zijn")
	}
}

// TestReset_Online controleert of de ontdekking wordt geactiveerd bij een 'online' status
func TestHADiscovery_Reset_Online(t *testing.T) {
	mockQueue := NewMockQueue()

	// We configureren een test-entity om te kijken of de JSON-generatie en publicatie werkt
	entity := HAEntityConfig{
		Component:  "sensor",
		Name:       "Temperatuur",
		UniqueId:   "test_temp_01",
		StateTopic: "pellet/state/temp",
		Device: HADevice{
			Identifiers:  []string{"test_id"},
			Name:         "Kachel",
			Manufacturer: "MicroNova",
			Model:        "Basic",
		},
	}

	discovery := HADiscovery{
		DeviceName: "Test Kachel",
		Entities:   []HAEntityConfig{entity},
		Publisher:  mockQueue,
	}

	// Activeer de reset methode met de 'online' payload
	discovery.reset(HAStatusTopic, HAOnline)

	var foundPayload string
	var foundTopic string
	for topic, payload := range mockQueue.Published {
		if strings.Contains(topic, entity.UniqueId) {
			foundTopic = topic
			foundPayload = payload
			break
		}
	}

	if foundTopic == "" {
		t.Fatalf("Geen ontdekkingspayload gepubliceerd voor unique_id: %s", entity.UniqueId)
	}

	// Valideer of de payload geldige JSON is en de correcte velden bevat
	var config HAEntityConfig
	err := json.Unmarshal([]byte(foundPayload), &config)
	if err != nil {
		t.Fatalf("Gepubliceerde payload is geen geldige JSON: %v", err)
	}

	if config.Name != entity.Name {
		t.Errorf("Verwachtte naam %s, maar kreeg %s", entity.Name, config.Name)
	}

	// Controleer dat het interne `Component` veld NIET in de JSON zit (vanwege json:"-")
	if strings.Contains(foundPayload, `"Component"`) || strings.Contains(foundPayload, `"component"`) {
		t.Error("Het interne veld 'Component' mag niet worden meegegeven in de JSON string")
	}
}

// TestReset_Offline controleert dat er niets gebeurt als Home Assistant 'offline' meldt
func TestHADiscovery_Reset_Offline(t *testing.T) {
	mockQueue := NewMockQueue()
	discovery := HADiscovery{
		DeviceName: "Test Device",
		Entities: []HAEntityConfig{
			{Component: "sensor", UniqueId: "test_sensor"},
		},
		Publisher: mockQueue,
	}

	// Stuur een 'offline' status
	discovery.reset(HAStatusTopic, "offline")

	// Er zouden geen berichten gepubliceerd moeten zijn
	if len(mockQueue.Published) > 0 {
		t.Errorf("Er zijn %d berichten gepubliceerd, maar verwachtte 0 bij status offline", len(mockQueue.Published))
	}
}
