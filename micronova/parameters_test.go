package micronova

import (
	"testing"
)

// --- Tests ---

func TestGetText_WithValueDescr(t *testing.T) {
	vd := []valueDescr{{Value: 5, Description: "five"}}
	txt := getText(5, "#+1", "{0}", vd)
	if txt != "five" {
		t.Fatalf("expected 'five', got %q", txt)
	}
}

func TestGetText_Formula(t *testing.T) {
	// formula: x*2, value 3 => 6
	txt := getText(3, "#*2", "{0}", nil)
	if txt != "6" && txt != "6.0" { // formula may produce "6" or "6.0" depending on formatter
		t.Fatalf("expected 6-like output, got %q", txt)
	}
}

func TestIsActive(t *testing.T) {
	parameters = []parameter{
		{regKey: "status_get", text: "Off"},
	}
	if isActive() {
		t.Fatal("expected inactive when status_get == Off")
	}
	parameters[0].text = "On"
	if !isActive() {
		t.Fatal("expected active when status_get != Off")
	}
}
