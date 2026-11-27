package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	deviceJobPath  = "deviceJobStatus/"
	retryDeviceJob = 30
)

type jobAnswerData struct {
	Items  []int  `json:"Items"`
	Values []int  `json:"Values"`
	Cmd    string `json:"cmd"`
}

type deviceJobResp struct {
	Success          bool           `json:"Success"`
	Error            string         `json:"Error"`
	Text             string         `json:"Text"`
	JobRequestStatus string         `json:"jobRequestStatus"`
	JobAnswerStatus  string         `json:"jobAnswerStatus"`
	JobAnswerData    *jobAnswerData `json:"jobAnswerData,omitempty"`
}

// handle empty jobAnswerData returned as an empty string by the API
func (d *jobAnswerData) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`""`)) {
		return nil
	}

	type Alias jobAnswerData
	return json.Unmarshal(data, (*Alias)(d))
}

func getJobResult(requestId string) error {
	for range retryDeviceJob {
		time.Sleep(time.Second)
		if deviceJob(requestId) {
			return nil
		}
	}
	return fmt.Errorf("No DeviceJob result within %d seconds", retryDeviceJob)
}

func deviceJob(requestId string) bool {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceJobRequest(ctx, requestId)
	if req == nil {
		log.Error().Msg("Failed to build DeviceJob request")
	}

	log.Debug().Msgf("API Call: %s%s", deviceJobPath, requestId)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("DeviceJob HTTP Client error")
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
	}

	log.Trace().Msgf("DeviceJob Response:\n%s", string(body[:]))

	var result deviceJobResp
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error().Err(err).Msg("Error unmarshaling DeviceJob response")
		return false
	}

	// keep trying if status is 'waiting' or 'terminated'
	if result.JobAnswerStatus != "completed" {
		return false
	}

	// keep trying if jobAnswerData is empty
	if result.JobAnswerData == nil {
		return false
	}

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		log.Error().Msgf("DeviceJob Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if result.JobAnswerData.Cmd == "GetBufferReading" {
		offsets = nil
		for i := 0; i < len(result.JobAnswerData.Items); i++ {
			offset := result.JobAnswerData.Items[i]
			offsets = append(offsets, offset)
		}
		log.Debug().Msgf("Offsets:\n%v", offsets)
	} else {
		for i := 0; i < len(result.JobAnswerData.Items); i++ {
			offset := result.JobAnswerData.Items[i]
			value := result.JobAnswerData.Values[i]

			for p := 0; p < len(parameters); p++ {
				if parameters[p].offset == offset {
					parameters[p].value = value
					parameters[p].realValue = getCalculated(value, parameters[p].formula)
					parameters[p].text = getText(value, parameters[p].formula, parameters[p].format, parameters[p].valueDescr)
				}
			}
		}
	}

	return true
}

func buildDeviceJobRequest(ctx context.Context, requestId string) *http.Request {
	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: deviceJobPath + requestId}
	log.Trace().Msgf("DeviceJob URL:\n%v", reqUrl.String())

	// Create Request
	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl.String(), nil)
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
