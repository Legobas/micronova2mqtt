package appdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- helpers -------------------------------------------------------------

func newDataManager(paths ...string) *DataManager {
	return &DataManager{paths: paths}
}

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, configFilename)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

const validYAML = `
mqtt:
  url: "tcp://localhost:1883"
micronova:
  brand: "michi"
  email: "user@example.com"
  password: "secret"
`

// --- validateConfig tests ----------------------------------------------

func TestValidateConfig_Valid(t *testing.T) {
	dm := newDataManager()
	if err := yaml.Unmarshal([]byte(validYAML), &dm.Config); err != nil {
		t.Fatal(err)
	}
	if err := dm.validateConfig(); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

func TestValidateConfig_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantMsg string
	}{
		{
			name:    "missing mqtt url",
			yaml:    "micronova:\n  brand: b\n  email: e\n  password: p\n",
			wantMsg: "MQTT Server URL is mandatory",
		},
		{
			name:    "missing brand",
			yaml:    "mqtt:\n  url: u\nmicronova:\n  email: e\n  password: p\n",
			wantMsg: "Micronova Brand is mandatory",
		},
		{
			name:    "missing email",
			yaml:    "mqtt:\n  url: u\nmicronova:\n  brand: b\n  password: p\n",
			wantMsg: "Micronova email address is mandatory",
		},
		{
			name:    "missing password",
			yaml:    "mqtt:\n  url: u\nmicronova:\n  brand: b\n  email: e\n",
			wantMsg: "Micronova password is mandatory",
		},
		{
			name:    "min greater than max",
			yaml:    "mqtt:\n  url: u\nmicronova:\n  brand: b\n  email: e\n  password: p\n  actions:\n    - trigger:\n        get_key: temp\n        min_value: 10\n        max_value: 5\n",
			wantMsg: "temp: Minimum Value must be lower or equal to Maximum value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := newDataManager()
			if err := yaml.Unmarshal([]byte(tt.yaml), &dm.Config); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			err := dm.validateConfig()
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantMsg)
			}
			if err.Error() != tt.wantMsg {
				t.Fatalf("got error %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestValidateConfig_MinEqualsMax_IsValid(t *testing.T) {
	dm := newDataManager()
	yamlStr := "mqtt:\n  url: u\nmicronova:\n  brand: b\n  email: e\n  password: p\n  actions:\n    - trigger:\n        get_key: temp\n        min_value: 5\n        max_value: 5\n"
	if err := yaml.Unmarshal([]byte(yamlStr), &dm.Config); err != nil {
		t.Fatal(err)
	}
	if err := dm.validateConfig(); err != nil {
		t.Fatalf("min == max should be valid, got: %v", err)
	}
}

// --- loadConfig tests ----------------------------------------------------

func TestLoadConfig_Success(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, validYAML)

	dm := newDataManager(path)
	if err := dm.loadConfig(); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if dm.configPath != path {
		t.Fatalf("configPath = %q, want %q", dm.configPath, path)
	}
	if dm.dataDir != dir {
		t.Fatalf("dataDir = %q, want %q", dm.dataDir, dir)
	}
	if dm.sessionPath != filepath.Join(dir, sessionFilename) {
		t.Fatalf("sessionPath = %q, want %q", dm.sessionPath, filepath.Join(dir, sessionFilename))
	}
	if dm.Config.Mqtt.Url != "tcp://localhost:1883" {
		t.Fatalf("MQTT url not parsed: %q", dm.Config.Mqtt.Url)
	}
	if dm.Config.Micronova.Brand != "michi" {
		t.Fatalf("brand not parsed: %q", dm.Config.Micronova.Brand)
	}
}

func TestLoadConfig_UsesFirstExistingPath(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	second := writeConfig(t, dir, validYAML+"\n# from second\n")

	dm := newDataManager(
		filepath.Join(sub, "nonexistent.yaml"), // missing, skipped
		second,                                 // exists -> wins
	)
	if err := dm.loadConfig(); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if dm.configPath != second {
		t.Fatalf("expected first existing path to win, got %q", dm.configPath)
	}
	if !strings.Contains(strings.Join(msgPaths(t, dm), "\n"), "") {
		// no-op; kept simple
	}
	if dm.dataDir != dir {
		t.Fatalf("dataDir = %q, want %q", dm.dataDir, dir)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.yaml")

	dm := newDataManager(missing)
	err := dm.loadConfig()
	if err == nil {
		t.Fatal("expected error when no config file exists")
	}
	if !strings.Contains(err.Error(), "configuration file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("error should list searched paths, got: %v", err)
	}
	if dm.configPath != "" {
		t.Fatalf("configPath should stay empty, got %q", dm.configPath)
	}
}

func TestLoadConfig_NoPaths(t *testing.T) {
	dm := newDataManager() // dm.paths is empty
	err := dm.loadConfig()
	if err == nil {
		t.Fatal("expected error when paths list is empty")
	}
	if !strings.Contains(err.Error(), "configuration file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, "mqtt: [not: valid")

	dm := newDataManager(path)
	err := dm.loadConfig()
	if err == nil {
		t.Fatal("expected unmarshal error for malformed YAML")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal config file") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error should mention the offending file, got: %v", err)
	}
}

func TestLoadConfig_InvalidContentFailsValidation(t *testing.T) {
	dir := t.TempDir()
	// Parseable YAML, but missing mandatory fields.
	path := writeConfig(t, dir, "mqtt:\n  url: u\n")

	dm := newDataManager(path)
	err := dm.loadConfig()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid config from") ||
		!strings.Contains(err.Error(), path) {
		t.Fatalf("expected wrapped validation error mentioning %q, got: %v", path, err)
	}
}

// drop if you don't need it:
func msgPaths(t *testing.T, dm *DataManager) []string {
	t.Helper()
	return nil
}
