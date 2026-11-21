package appdata

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	fileMode os.FileMode = 0600
)

type Session struct {
	UUID         string `yaml:"uuid"`
	Token        string `yaml:"tkn"`
	RefreshToken string `yaml:"rtn"`
	ProductId    string `yaml:"pid"`
	DeviceId     string `yaml:"did"`
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
