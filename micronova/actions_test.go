package micronova

import (
	"micronova2mqtt/files"
	"testing"
)

// MockParameter represents a mock parameter for testing
type MockParameter struct {
	regKey string
	value  int
}

// TestActionsHolder_Reset tests the reset functionality
func TestActionsHolder_Reset(t *testing.T) {
	activated = []string{"test=1", "another=2"}

	holder := &ActionsHolder{}
	holder.reset()

	if len(activated) != 0 {
		t.Errorf("reset() failed: expected activated to be empty, got %v", activated)
	}
}

// TestActionsHolder_Process_ActionActivated tests basic action activation
func TestActionsHolder_Process_ActionActivated(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Mock parameters
	originalParams := parameters
	parameters = []parameter{
		{regKey: "temp_sensor", value: 25},
	}
	defer func() { parameters = originalParams }()

	// Create test action
	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "temp_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// Verify action was activated
	if len(activated) != 1 {
		t.Errorf("Process() failed: expected 1 activated action, got %d", len(activated))
	}
	if activated[0] != "temp_sensor=20" {
		t.Errorf("Process() failed: expected 'temp_sensor=20', got '%s'", activated[0])
	}
}

// TestActionsHolder_Process_ActionNotActivated tests when real value is below trigger
func TestActionsHolder_Process_ActionNotActivated(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Mock parameters
	originalParams := parameters
	parameters = []parameter{
		{regKey: "temp_sensor", value: 15},
	}
	defer func() { parameters = originalParams }()

	// Create test action with higher trigger value
	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "temp_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// Verify action was not activated
	if len(activated) != 0 {
		t.Errorf("Process() failed: expected 0 activated actions, got %d", len(activated))
	}
}

// TestActionsHolder_Process_DuplicateActivation tests that duplicate actions are skipped
func TestActionsHolder_Process_DuplicateActivation(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Pre-populate activated
	activated = []string{"temp_sensor=20"}

	// Mock parameters
	originalParams := parameters
	parameters = []parameter{
		{regKey: "temp_sensor", value: 25},
	}
	defer func() { parameters = originalParams }()

	// Create same test action
	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "temp_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// Verify action was not activated again
	if len(activated) != 1 {
		t.Errorf("Process() failed: expected 1 activated action (no duplicates), got %d", len(activated))
	}
}

// TestActionsHolder_Process_MultipleActions tests processing multiple actions
func TestActionsHolder_Process_MultipleActions(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Mock parameters
	originalParams := parameters
	parameters = []parameter{
		{regKey: "temp_sensor", value: 25},
		{regKey: "humidity_sensor", value: 80},
	}
	defer func() { parameters = originalParams }()

	// Create multiple test actions
	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "temp_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
		{
			Trigger: files.ActionTrigger{
				GetKey:       "humidity_sensor",
				MinimumValue: 70,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "dehumidifier", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// Verify both actions were activated
	if len(activated) != 2 {
		t.Errorf("Process() failed: expected 2 activated actions, got %d", len(activated))
	}
}

// TestActionsHolder_Process_EdgeCase_EqualValue tests when real value equals trigger value
func TestActionsHolder_Process_EdgeCase_EqualValue(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Mock parameters - value equals minimum
	originalParams := parameters
	parameters = []parameter{
		{regKey: "temp_sensor", value: 20},
	}
	defer func() { parameters = originalParams }()

	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "temp_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// Verify action was activated (>= condition)
	if len(activated) != 1 {
		t.Errorf("Process() failed: expected 1 activated action, got %d", len(activated))
	}
}

// TestActionsHolder_Process_UnknownTriggerKey tests when trigger key doesn't exist in parameters
func TestActionsHolder_Process_UnknownTriggerKey(t *testing.T) {
	// Setup
	holder := &ActionsHolder{}
	holder.reset()

	// Mock parameters without the trigger key
	originalParams := parameters
	parameters = []parameter{
		{regKey: "other_sensor", value: 25},
	}
	defer func() { parameters = originalParams }()

	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "unknown_sensor",
				MinimumValue: 20,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "heater", Value: 1},
			},
		},
	}

	holder.Process(actions)

	// With realValue = 0, it should not be activated (0 < 20)
	if len(activated) != 0 {
		t.Errorf("Process() failed: expected 0 activated actions for unknown key, got %d", len(activated))
	}
}
