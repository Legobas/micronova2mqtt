package micronova

import (
	"micronova2mqtt/files"
	"testing"
)

func TestActionsHolder_Reset(t *testing.T) {
	holder := &ActionsHolder{
		activated: []string{"temp >= 20.0 & <= 25.0"},
	}

	holder.reset()

	if len(holder.activated) != 0 {
		t.Errorf("Expected empty activated list after reset, got %v", holder.activated)
	}
}

func TestActionsHolder_Process_ActivatesAction(t *testing.T) {
	// Arrange
	mockSetParameters := make(map[string]int)

	// Create the slice and pass its pointer
	testParameters := []parameter{
		{regKey: "room_temp", realValue: 22, formula: "none"},
	}

	holder := &ActionsHolder{
		parameters: &testParameters, // Pointer to slice
		getCalculated: func(value int, formulaStr string) float64 {
			return float64(value)
		},
		setParameterByRegKey: func(key string, value int) {
			mockSetParameters[key] = value
		},
	}

	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "room_temp",
				MinimumValue: 20.0,
				MaximumValue: 25.0,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "stove_power", Value: 3},
			},
		},
	}

	// Act
	holder.Process(actions)

	// Assert
	expectedAct := "room_temp >= 20.0 & <= 25.0"
	if len(holder.activated) != 1 || holder.activated[0] != expectedAct {
		t.Errorf("Expected action '%s' in activated list, got %v", expectedAct, holder.activated)
	}

	if val, ok := mockSetParameters["stove_power"]; !ok || val != 3 {
		t.Errorf("Expected stove_power to be set to 3, got %v", val)
	}
}

func TestActionsHolder_Process_SkipsOutOfRange(t *testing.T) {
	// Arrange
	mockSetParameters := make(map[string]int)

	testParameters := []parameter{
		{regKey: "room_temp", value: 18, formula: "none"}, // 18 is out of range [20.0, 25.0]
	}

	holder := &ActionsHolder{
		parameters: &testParameters,
		getCalculated: func(value int, formulaStr string) float64 {
			return float64(value)
		},
		setParameterByRegKey: func(key string, value int) {
			mockSetParameters[key] = value
		},
	}

	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "room_temp",
				MinimumValue: 20.0,
				MaximumValue: 25.0,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "stove_power", Value: 3},
			},
		},
	}

	// Act
	holder.Process(actions)

	// Assert
	if len(holder.activated) != 0 {
		t.Errorf("Expected no activated actions for out-of-range value, got %v", holder.activated)
	}

	if len(mockSetParameters) != 0 {
		t.Errorf("Expected no parameter changes, but got: %v", mockSetParameters)
	}
}

func TestActionsHolder_Process_AlreadyActivated(t *testing.T) {
	// Arrange
	mockSetParameters := make(map[string]int)
	triggerAct := "room_temp >= 20.0 & <= 25.0"

	testParameters := []parameter{
		{regKey: "room_temp", value: 22, formula: "none"},
	}

	holder := &ActionsHolder{
		activated:  []string{triggerAct}, // Already activated previously
		parameters: &testParameters,
		getCalculated: func(value int, formulaStr string) float64 {
			return float64(value)
		},
		setParameterByRegKey: func(key string, value int) {
			mockSetParameters[key] = value
		},
	}

	actions := []files.Action{
		{
			Trigger: files.ActionTrigger{
				GetKey:       "room_temp",
				MinimumValue: 20.0,
				MaximumValue: 25.0,
			},
			SetValues: []files.ActionSetValues{
				{SetKey: "stove_power", Value: 3},
			},
		},
	}

	// Act
	holder.Process(actions)

	// Assert
	if len(mockSetParameters) != 0 {
		t.Error("Action was already activated; setParameterByRegKey should not have been called")
	}
}
