package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	deviceInfoPath = "deviceGetInfo"
)

type deviceInfoData struct {
	Name                       string    `json:"name"`
	Description                string    `json:"description"`
	IdProduct                  string    `json:"id_product"`
	Mac                        string    `json:"mac"`
	Online                     bool      `json:"is_online"`
	AssistanceCode             string    `json:"assistance_code"`
	SecurityCode               string    `json:"security_code"`
	CreationDateDevice         time.Time `json:"creation_date_device"`
	UpdateDateDevice           time.Time `json:"update_date_device"`
	CustomerCode               string    `json:"customer_code"`
	CodArt                     string    `json:"cod_art"`
	Serial                     string    `json:"serial"`
	EnableChronoWeek           bool      `json:"enable_chrono_week"`
	EnableSetChronoWeek        bool      `json:"enable_set_chrono_week"`
	ChronoPrograms             int       `json:"chrono_programs"`
	SetPowerMin                int       `json:"set_power_min"`
	SetPowerMax                int       `json:"set_power_max"`
	EnableChronoSetTemperature bool      `json:"enable_chrono_set_temperature"`
	SetTemperatureMin          int       `json:"set_temperature_min"`
	SetTemperatureMax          int       `json:"set_temperature_max"`
	SetTemperatureWaterMin     int       `json:"set_temperature_water_min"`
	SetTemperatureWaterMax     int       `json:"set_temperature_water_max"`
	EnableChronoSetPower       bool      `json:"enable_chrono_set_power"`
	EnableMainVentilation      bool      `json:"enable_main_ventilation"`
	EnableCanalization1        bool      `json:"enable_canalization_1"`
	EnableCanalization2        bool      `json:"enable_canalization_2"`
	SetCanalization1Min        int       `json:"set_canalization_1_min"`
	SetCanalization1Max        int       `json:"set_canalization_1_max"`
	SetCanalization2Min        int       `json:"set_canalization_2_min"`
	SetCanalization2Max        int       `json:"set_canalization_2_max"`
	EnableWeekDays             bool      `json:"enable_week_days"`
	EnableWaterPuffer          bool      `json:"enable_water_puffer"`
	EnableWaterBoiler          bool      `json:"enable_water_boiler"`
	EnableWater                bool      `json:"enable_water"`
	EnableChronoSetFan         bool      `json:"enable_chrono_set_fan"`
	SetFanMin                  int       `json:"set_fan_min"`
	SetFanMax                  int       `json:"set_fan_max"`
	IdRegistersMap             string    `json:"id_registers_map"`
	RowVersion                 string    `json:"rowVersion"`
}

type deviceInfoResp struct {
	Success    bool             `json:"Success"`
	Text       string           `json:"Text"`
	Value      bool             `json:"Value"`
	DeviceInfo []deviceInfoData `json:"device_info"`
	Device     []device         `json:"device"`
	Error      string           `json:"Error"`
}

func deviceInfo() {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceInfoRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build DeviceInfo request")
	}

	log.Info().Msgf("API Call: %s", deviceInfoPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("DeviceInfo HTTP Client error")
	}
	defer resp.Body.Close()

	var result deviceInfoResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	log.Trace().Msgf("DeviceInfo Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("DeviceInfo Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if len(result.DeviceInfo) == 0 {
		log.Fatal().Msg("DeviceInfo: No DeviceInfo")
	}

	if len(result.Device) == 0 {
		log.Fatal().Msg("DeviceInfo: No Device")
	}

	for _, device := range result.Device {
		if device.ProductId == dm.Session.ProductId && device.DeviceId == dm.Session.DeviceId {
			if device.Online {
				state = stateInactive
			} else {
				log.Warn().Msgf("Device '%s' is offline", device.Name)
				state = stateOffline
			}

			deviceName = device.Name
			publisher(device.Name, "description", device.Description, true)
			publisher(device.Name, "product", device.ProductName, true)
			publisher(device.Name, "creationDate", fmt.Sprintf("%v", device.CreationDate), true)
			publisher(device.Name, "online", fmt.Sprintf("%v", device.Online), true)

			break
		}
	}
}

func buildDeviceInfoRequest(ctx context.Context) *http.Request {
	jsonData := []byte(`{"id_device": "` + dm.Session.DeviceId + `","id_product": "` + dm.Session.ProductId + `"}`)

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: deviceInfoPath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("DeviceInfo Request:\n%v", string(jsonData))

	// Create Request
	req, err := http.NewRequestWithContext(ctx, "POST", reqUrl.String(), payload)
	if err != nil {
		log.Error().Err(err).Msg("HTTP Request error")
		return nil
	}

	// Set HTTP headers
	req.Header.Set("Accept", acceptHeader)
	req.Header.Set("Content-Type", contentTypeHeader)
	req.Header.Set("id_brand", brandIdHeader)
	req.Header.Set("customer_code", customerCode)
	req.Header.Set("local", "false")
	req.Header.Set("Authorization", dm.Session.Token)

	return req
}
