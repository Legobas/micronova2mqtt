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
	readDevicePath = "deviceRequestReading"
)

type readDeviceReq struct {
	ProductId string `json:"id_product"`
	DeviceId  string `json:"id_device"`
	Protocol  string `json:"Protocol"`
	BitData   int    `json:"BitData"`
	Endianess string `json:"Endianess"`
	Freq      int    `json:"Freq"`
	Items     []int  `json:"Items"`
	Masks     []int  `json:"Masks"`
}

type readDeviceResp struct {
	Success   bool   `json:"Success"`
	Text      string `json:"Text"`
	RequestId string `json:"idRequest"`
	Error     string `json:"Error"`
}

func readDevice() {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildReadDeviceRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build ReadDevice request")
	}

	log.Info().Msgf("API Call: %s -->", readDevicePath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("ReadDevice HTTP Client error")
		return
	}
	defer resp.Body.Close()

	var result readDeviceResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	log.Trace().Msgf("ReadDevice Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("ReadDevice Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if result.RequestId == "" {
		log.Fatal().Msg("readDevice: No Request ID")
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

func buildReadDeviceRequest(ctx context.Context) *http.Request {
	items, masks := getParameters()
	registers := readDeviceReq{
		ProductId: dm.Session.ProductId,
		DeviceId:  dm.Session.DeviceId,
		Protocol:  "RWMSmaster",
		BitData:   8,
		Endianess: "L",
		Freq:      0,
		Items:     items,
		Masks:     masks,
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(registers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: readDevicePath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("ReadDevice Request:\n%v", string(jsonData))

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
