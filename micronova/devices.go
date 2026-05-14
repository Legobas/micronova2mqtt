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
	devicesPath = "deviceList"
)

type device struct {
	Id            string    `json:"id"`
	DeviceId      string    `json:"id_device"`
	ProductId     string    `json:"id_product"`
	ClientId      string    `json:"id_client"`
	AppId         string    `json:"id_app"`
	ProductSerial string    `json:"product_serial"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CreationDate  time.Time `json:"creation_date"`
	UpdateDate    time.Time `json:"update_date"`
	Deleted       bool      `json:"deleted"`
	Online        bool      `json:"is_online"`
	Mac           string    `json:"mac"`
	SecurityCode  string    `json:"security_code"`
	ProductName   string    `json:"name_product"`
	Serial        string    `json:"serial"`
}

type devicesResp struct {
	Success bool     `json:"Success"`
	Text    string   `json:"Text"`
	Devices []device `json:"device"`
	Error   string   `json:"Error"`
}

func deviceList() {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceListRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build DeviceList request")
	}

	log.Info().Msgf("API Call: %s", devicesPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("DeviceList HTTP Client error")
		return
	}
	defer resp.Body.Close()

	var result devicesResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	log.Trace().Msgf("DeviceList Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("DeviceList Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if len(result.Devices) == 0 {
		log.Fatal().Msg("DeviceList: No Devices")
	}

	for _, device := range result.Devices {
		if !device.Deleted {
			if err := dm.SetDeviceIds(device.ProductId, device.DeviceId); err != nil {
				log.Error().Err(err).Msg("Failed to set session Product/Device ID")
			}
			if !device.Online {
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

func buildDeviceListRequest(ctx context.Context) *http.Request {
	jsonData := []byte(`{}`)

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: devicesPath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("DeviceList Request:\n%v", string(jsonData))

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
