package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

const (
	writeDevicePath = "deviceRequestWriting"
)

type writeDeviceReq struct {
	ProductId string `json:"id_product"`
	DeviceId  string `json:"id_device"`
	Protocol  string `json:"Protocol"`
	BitData   int    `json:"BitData"`
	Endianess string `json:"Endianess"`
	Items     []int  `json:"Items"`
	Values    []int  `json:"Values"`
	Masks     []int  `json:"Masks"`
}

type writedeviceResp struct {
	Success   bool   `json:"Success"`
	Text      string `json:"Text"`
	RequestId string `json:"idRequest"`
	Error     string `json:"Error"`
}

func writeDevice(item int, value int, mask int) {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildWriteDeviceRequest(ctx, item, value, mask)
	if req == nil {
		log.Fatal().Msg("Failed to build WriteDevice request")
	}

	log.Info().Msgf("API Call: %s -->", writeDevicePath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("WriteDevice HTTP request failed")
		return
	}
	defer resp.Body.Close()

	var result writedeviceResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	log.Trace().Msgf("WriteDevice Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("WriteDevice Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if result.RequestId == "" {
		log.Fatal().Msg("WriteDevice: No Request ID")
	}

	if err := getJobResult(result.RequestId); err != nil {
		state = stateNotResponding
		log.Error().Err(err).Msg("Device is not responding")
		return
	}

	for _, par := range parameters {
		log.Trace().Msgf("Parameter %s=%v", par.regKey, par.text)
	}
}

func buildWriteDeviceRequest(ctx context.Context, item int, value int, mask int) *http.Request {
	registers := writeDeviceReq{
		ProductId: dm.Session.ProductId,
		DeviceId:  dm.Session.DeviceId,
		Protocol:  "RWMSmaster",
		BitData:   8,
		Endianess: "L",
		Items:     []int{item},
		Values:    []int{value},
		Masks:     []int{mask},
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(registers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: writeDevicePath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("WriteDevice Request: \n%v", string(jsonData))

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
