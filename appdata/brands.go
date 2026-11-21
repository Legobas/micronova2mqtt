package appdata

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type BrandsList map[string]struct {
	AppName      string `yaml:"app-name"`
	CustomerCode string `yaml:"customer-code"`
	Domain       string `yaml:"domain"`
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
