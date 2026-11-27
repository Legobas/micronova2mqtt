package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

const (
	signupPath    = "appSignup"
	PhoneType     = "Android"
	PhoneVersion  = "1.0"
	PhoneLanguage = "en"
)

type signupReq struct {
	PhoneType              string `json:"phone_type"`
	PhoneId                string `json:"phone_id"`
	PhoneVersion           string `json:"phone_version"`
	Language               string `json:"language"`
	AppId                  string `json:"id_app"`
	PushNotificationToken  string `json:"push_notification_token"`
	PushNotificationActive bool   `json:"push_notification_active"`
}

type signupResp struct {
	Success bool   `json:"Success"`
	Text    string `json:"Text"`
	AppId   string `json:"id_app"`
	Error   string `json:"Error"`
}

func signup() error {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildSignupRequest(ctx)
	if req == nil {
		err := fmt.Errorf("Failed to build Signup request")
		log.Error().Err(err)
		return err
	}

	log.Debug().Msgf("API Call: %s", signupPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Signup HTTP Client error")
		return err
	}
	defer resp.Body.Close()

	var result signupResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding response")
		return err
	}

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusCreated || !result.Success {
		err := fmt.Errorf("Signup Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
		log.Error().Err(err)
		return err
	}

	if result.AppId != dm.Session.UUID {
		log.Warn().Msg("Signup: AppId does not equal session UUID")
	}

	log.Info().Msgf("UUID Registered: %s", dm.Session.UUID)

	// Save session
	if err := dm.WriteSession(); err != nil {
		log.Error().Err(err).Msg("Failed to write session")
		return err
	}

	return nil
}

func buildSignupRequest(ctx context.Context) *http.Request {
	uuid := dm.CreateUUID()
	signup := signupReq{
		PhoneType:              PhoneType,
		PhoneId:                uuid,
		PhoneVersion:           PhoneVersion,
		Language:               PhoneLanguage,
		AppId:                  uuid,
		PushNotificationToken:  uuid,
		PushNotificationActive: false,
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(signup)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: signupPath}
	payload := bytes.NewReader(jsonData)
	log.Trace().Msgf("Signup Request:\n%v", string(jsonData))

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

	return req
}
