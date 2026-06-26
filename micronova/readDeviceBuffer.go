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
	readDeviceBufferPath = "deviceGetBufferReading"
)

type readDeviceBufferReq struct {
	ProductId string `json:"id_product"`
	DeviceId  string `json:"id_device"`
	BufferId  int    `json:"BufferId"`
}

func readDeviceBuffer() {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildReadDeviceBufferRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build ReadDeviceBuffer request")
	}

	log.Info().Msgf("API Call: %s -->", readDeviceBufferPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("ReadDeviceBuffer HTTP Client error")
		return
	}
	defer resp.Body.Close()

	var result readDeviceResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	log.Trace().Msgf("ReadDeviceBuffer Response:\n%+v", result)

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Fatal().Msgf("ReadDeviceBuffer Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if result.RequestId == "" {
		log.Fatal().Msg("readDeviceBuffer: No Request ID")
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

func buildReadDeviceBufferRequest(ctx context.Context) *http.Request {
	registers := readDeviceBufferReq{
		ProductId: dm.Session.ProductId,
		DeviceId:  dm.Session.DeviceId,
		BufferId:  1,
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(registers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: readDeviceBufferPath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("ReadDeviceBuffer Request:\n%v", string(jsonData))

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
