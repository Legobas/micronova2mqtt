package micronova

import (
	"testing"

	"micronova2mqtt/files"
)

func TestProcessActivatesActionWhenValueIsInRange(t *testing.T) {
	oldParameters, oldActivated := parameters, activated
	t.Cleanup(func() {
		parameters, activated = oldParameters, oldActivated
	})

	parameters = []parameter{
		{regKey: "trigger", value: 10},
		{regKey: "target", value: 0},
	}
	activated = nil

	holder := &ActionsHolder{}
	holder.Process([]files.Action{{
		Trigger: files.ActionTrigger{
			GetKey:       "trigger",
			MinimumValue: 5,
			MaximumValue: 15,
		},
	}})

	if len(activated) != 1 {
		t.Fatalf("activated count = %d, want 1", len(activated))
	}
}

func TestProcessDoesNotActivateActionOutsideRange(t *testing.T) {
	oldParameters, oldActivated := parameters, activated
	t.Cleanup(func() {
		parameters, activated = oldParameters, oldActivated
	})

	parameters = []parameter{
		{regKey: "trigger", value: 20},
		{regKey: "target", value: 0},
	}
	activated = nil

	holder := &ActionsHolder{}
	holder.Process([]files.Action{{
		Trigger: files.ActionTrigger{
			GetKey:       "trigger",
			MinimumValue: 5,
			MaximumValue: 15,
		},
		SetValues: []files.ActionSetValues{
			{SetKey: "target", Value: 42},
		},
	}})

	if got := valueForKey("target"); got != 0 {
		t.Fatalf("target value = %d, want 0", got)
	}
	if len(activated) != 0 {
		t.Fatalf("activated count = %d, want 0", len(activated))
	}
}

func TestProcessDoesNotActivateSameTriggerTwice(t *testing.T) {
	oldParameters, oldActivated := parameters, activated
	t.Cleanup(func() {
		parameters, activated = oldParameters, oldActivated
	})

	parameters = []parameter{
		{regKey: "trigger", value: 10},
		{regKey: "target", value: 0},
	}
	activated = nil

	action := files.Action{
		Trigger: files.ActionTrigger{
			GetKey:       "trigger",
			MinimumValue: 5,
			MaximumValue: 15,
		},
		SetValues: []files.ActionSetValues{
			{SetKey: "target", Value: 42},
		},
	}

	holder := &ActionsHolder{}
	holder.Process([]files.Action{action})

	// A repeated Process call should skip this trigger after activation.
	setParameterForTest("target", 7)
	holder.Process([]files.Action{action})

	if got := valueForKey("target"); got != 7 {
		t.Fatalf("target value = %d, want 7; action ran again", got)
	}
	if len(activated) != 1 {
		t.Fatalf("activated count = %d, want 1", len(activated))
	}
}

func valueForKey(key string) int {
	for _, p := range parameters {
		if p.regKey == key {
			return p.value
		}
	}
	return 0
}

func setParameterForTest(key string, value int) {
	for i := range parameters {
		if parameters[i].regKey == key {
			parameters[i].value = value
			return
		}
	}
}
