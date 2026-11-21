package appdata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Mqtt struct {
	Url       string `yaml:"url"`
	Username  string `yaml:"username"`   // default = ""
	Password  string `yaml:"password"`   // default = ""
	Qos       int    `yaml:"qos"`        // default = 0 (AtMostOnce)
	Retain    bool   `yaml:"retain"`     // default = false
	BaseTopic string `yaml:"base_topic"` // default = ""
}

type Power struct {
	On  string `yaml:"on"`
	Off string `yaml:"off"`
}

type ActionTrigger struct {
	GetKey       string  `yaml:"get_key"`
	MinimumValue float64 `yaml:"min_value"`
	MaximumValue float64 `yaml:"max_value"`
}

type ActionSetValues struct {
	SetKey string `yaml:"set_key"`
	Value  int    `yaml:"value"`
}

type Action struct {
	Trigger   ActionTrigger     `yaml:"trigger"`
	SetValues []ActionSetValues `yaml:"set_values"`
}

type RegKey struct {
	Key   string `yaml:"key"`
	Title string `yaml:"title"`
}

type Micronova struct {
	Brand      string   `yaml:"brand"`
	Email      string   `yaml:"email"`
	Password   string   `yaml:"password"`
	Power      Power    `yaml:"power"`
	LogRegKeys bool     `yaml:"log_reg_keys"`
	Actions    []Action `yaml:"actions"`
	RegKeys    []RegKey `yaml:"reg_keys"`
}

type HomeAssistant struct {
	Discovery bool `yaml:"discovery"`
}

type Config struct {
	Mqtt          Mqtt          `yaml:"mqtt"`
	Micronova     Micronova     `yaml:"micronova"`
	HomeAssistant HomeAssistant `yaml:"homeassistant"`
}

func (dm *DataManager) loadConfig() error {
	var msg string
	for _, path := range dm.paths {
		msg += path + " "
		if data, err := os.ReadFile(path); err == nil {
			err = yaml.Unmarshal(data, &dm.Config)
			if err != nil {
				return fmt.Errorf("failed to unmarshal config file: %s: %w", path, err)
			}
			dm.configPath = path
			dm.dataDir = filepath.Dir(path)
			dm.sessionPath = filepath.Join(dm.dataDir, sessionFilename)
			break
		}
	}

	if len(dm.configPath) == 0 {
		return fmt.Errorf("configuration file not found in paths: %s", msg)
	}

	if len(dm.dataDir) == 0 {
		return errors.New("data dir is empty")
	}

	if err := dm.validateConfig(); err != nil {
		return fmt.Errorf("invalid config from %s: %w", dm.configPath, err)
	}

	log.Debug().Msgf("Configuration loaded: %s", dm.configPath)
	// Trace logging is disabled for security reasons
	// log.Trace().Msgf("MQTT Config: %+v", dm.config.Mqtt)
	// log.Trace().Msgf("Micronova Config: %+v", dm.config.Micronova)

	return nil
}

func (dm DataManager) validateConfig() error {
	if dm.Config.Mqtt.Url == "" {
		return errors.New("MQTT Server URL is mandatory")
	}
	if dm.Config.Micronova.Brand == "" {
		return errors.New("Micronova Brand is mandatory")
	}
	if dm.Config.Micronova.Email == "" {
		return errors.New("Micronova email address is mandatory")
	}
	if dm.Config.Micronova.Password == "" {
		return errors.New("Micronova password is mandatory")
	}
	for _, action := range dm.Config.Micronova.Actions {
		if action.Trigger.MinimumValue > action.Trigger.MaximumValue {
			return errors.New(action.Trigger.GetKey + ": Minimum Value must be lower or equal to Maximum value")
		}
	}
	return nil
}
