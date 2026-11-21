package appdata

import (
	"os"
	"path/filepath"
)

const (
	dataDir         = "data"
	topicBase       = "micronova2mqtt"
	configFilename  = "micronova2mqtt.yml"
	sessionFilename = "session.dat"
	brandsFilename  = "brands.yml"
)

type DataManager struct {
	Config      Config
	Session     Session
	configPath  string
	sessionPath string
	dataDir     string
	paths       []string
}

func NewDataManager() (*DataManager, error) {
	var err error

	dm := &DataManager{}
	dm.paths = getPaths()

	// Load configuration
	if err = dm.loadConfig(); err != nil {
		return nil, err
	}

	// Load session (Non-fatal if missing)
	dm.loadSession()

	if dm.Config.Mqtt.BaseTopic == "" {
		dm.Config.Mqtt.BaseTopic = topicBase
	}

	return dm, nil
}

func getPaths() []string {
	var paths []string
	paths = append(paths, filepath.Join("/", dataDir, configFilename))

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, dataDir, configFilename))
	}

	if workDir, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(workDir, configFilename))
		paths = append(paths, filepath.Join(workDir, dataDir, configFilename))
	}

	return paths
}
