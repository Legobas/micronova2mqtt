package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	dataDir                     = "data"
	configFilename              = "micronova2mqtt.yml"
	sessionFilename             = "session.dat"
	brandsFilename              = "brands.yml"
	fileMode        os.FileMode = 0600
)

type Mqtt struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"` // default = ""
	Password string `yaml:"password"` // default = ""
	Qos      int    `yaml:"qos"`      // default = 0 (AtMostOnce)
	Retain   bool   `yaml:"retain"`   // default = false
}

type Power struct {
	On  string `yaml:"on"`
	Off string `yaml:"off"`
}

type RegKey struct {
	Key   string `yaml:"key"`
	Topic string `yaml:"topic"`
}

type Micronova struct {
	Brand    string   `yaml:"brand"`
	Email    string   `yaml:"email"`
	Password string   `yaml:"password"`
	Power    Power    `yaml:"power"`
	RegKeys  []RegKey `yaml:"reg_keys"`
}

type Config struct {
	Mqtt      Mqtt      `yaml:"mqtt"`
	Micronova Micronova `yaml:"micronova"`
}

type Session struct {
	UUID         string `yaml:"uuid"`
	Token        string `yaml:"tkn"`
	RefreshToken string `yaml:"rtn"`
	ProductId    string `yaml:"pid"`
	DeviceId     string `yaml:"did"`
}

type BrandsList map[string]struct {
	AppName      string `yaml:"app-name"`
	CustomerCode string `yaml:"customer-code"`
	Domain       string `yaml:"domain"`
}

type DataManager struct {
	Config      Config
	Session     Session
	configPath  string
	sessionPath string
	dataDir     string
}

func NewData() (*DataManager, error) {
	dm := &DataManager{}

	// Load configuration
	if err := dm.loadConfig(); err != nil {
		return nil, err
	}

	// Load session
	dm.loadSession() // Non-fatal if missing

	return dm, nil
}

func (dm *DataManager) loadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.New("could not get Home Dir")
	}
	workDir, err := os.Getwd()
	if err != nil {
		return errors.New("could not get Work Dir")
	}

	paths := []string{
		filepath.Join("/", dataDir, configFilename),
		filepath.Join(homeDir, dataDir, configFilename),
		filepath.Join(workDir, configFilename),
		filepath.Join(workDir, dataDir, configFilename),
	}

	var msg string
	for _, path := range paths {
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
	return nil
}

func (dm *DataManager) loadSession() {
	data, err := os.ReadFile(dm.sessionPath)
	if err != nil {
		log.Debug().Msgf("Session file not found: %s", dm.sessionPath)
		return
	}
	decrypted, err := dm.decrypt(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to encrypt session")
	}

	if err := yaml.Unmarshal(decrypted, &dm.Session); err != nil {
		log.Error().Err(err).Msg("Could not unmarshal session")
		return
	}

	log.Debug().Msgf("Session file: %s", dm.sessionPath)
	// Trace logging is disabled for security reasons
	// log.Trace().Msgf("Session: %+v", dm.Session)
}

func (dm DataManager) WriteSession() error {
	sessionData, err := yaml.Marshal(dm.Session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	encrypted, err := dm.encrypt(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session: %w", err)
	}
	if err := os.WriteFile(dm.sessionPath, encrypted, fileMode); err != nil {
		return fmt.Errorf("failed to write session file %s: %w", dm.sessionPath, err)
	}

	log.Info().Msg("Session stored")
	return nil
}

func (dm DataManager) GetBrand(brand string) (customerCode, domain string, err error) {
	brandsFile := filepath.Join(dm.dataDir, brandsFilename)
	data, err := os.ReadFile(brandsFile)
	if err != nil {
		execDir, err := os.Executable()
		if err != nil {
			return "", "", fmt.Errorf("brands file not found")
		}
		execDir = filepath.Dir(execDir)
		brandsFile = filepath.Join(execDir, brandsFilename)
		data, err = os.ReadFile(brandsFile)
		if err != nil {
			return "", "", fmt.Errorf("could not read brands file %s", brandsFile)
		}
	}
	log.Debug().Msgf("brands file: %s", brandsFile)

	var brands BrandsList
	if err := yaml.Unmarshal(data, &brands); err != nil {
		return "", "", fmt.Errorf("error unmarshaling brands file: %w", err)
	}

	b, exists := brands[brand]
	if !exists {
		return "", "", fmt.Errorf("brand '%s' not found in brands file", brand)
	}

	log.Info().Msgf("Brand: %s, app: %s, customerCode: %s, domain: %s", brand, b.AppName, b.CustomerCode, b.Domain)

	return b.CustomerCode, b.Domain, nil
}

func (dm *DataManager) CreateUUID() string {
	dm.Session.UUID = uuid.NewString()
	log.Debug().Msgf("New UUID: %s", dm.Session.UUID)
	if err := dm.WriteSession(); err != nil {
		log.Error().Err(err).Msg("failed to write session")
	}
	return dm.Session.UUID
}

// Setters
func (dm *DataManager) SetSessionToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	dm.Session.Token = token

	log.Info().Msg("Session token set")
	if err := dm.WriteSession(); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	return nil
}

func (dm *DataManager) SetSessionRefreshToken(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}

	dm.Session.RefreshToken = refreshToken

	log.Info().Msg("Session refresh token set")
	if err := dm.WriteSession(); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	return nil
}

func (dm *DataManager) SetSessionTokens(token, refreshToken string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	if refreshToken == "" {
		return errors.New("refresh token cannot be empty")
	}

	dm.Session.Token = token
	dm.Session.RefreshToken = refreshToken

	log.Info().Msg("Session tokens set")
	if err := dm.WriteSession(); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	return nil
}

func (dm *DataManager) SetDeviceIds(productId, deviceId string) error {
	if productId == "" {
		return errors.New("productId cannot be empty")
	}
	if deviceId == "" {
		return errors.New("deviceId cannot be empty")
	}

	dm.Session.ProductId = productId
	dm.Session.DeviceId = deviceId

	if err := dm.WriteSession(); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}
	return nil
}
