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
	loginPath = "userLogin"
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginResp struct {
	Success      bool   `json:"Success"`
	Text         string `json:"Text"`
	Error        string `json:"Error"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func login() error {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildLoginRequest(ctx)
	if req == nil {
		err := fmt.Errorf("Failed to build Login request")
		log.Error().Err(err)
		return err
	}

	log.Debug().Msgf("API Call: %s", loginPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("DeviceList HTTP Client error")
		return err
	}
	defer resp.Body.Close()

	var result loginResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Error().Err(err).Msg("Error decoding response")
		return err
	}

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusOK || !result.Success {
		err := fmt.Errorf("Login Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
		log.Error().Err(err)
		return err
	}

	if result.Token == "" || result.RefreshToken == "" {
		err := fmt.Errorf("Login: Tokens missing")
		log.Error().Err(err)
		return err
	}

	if err := dm.SetSessionTokens(result.Token, result.RefreshToken); err != nil {
		log.Error().Err(err).Msg("Failed to set tokens")
		return err
	}

	return nil
}

func buildLoginRequest(ctx context.Context) *http.Request {
	login := loginReq{
		Email:    dm.Config.Micronova.Email,
		Password: dm.Config.Micronova.Password,
	}

	// Marshal JSON for request
	jsonData, err := json.Marshal(login)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request")
		return nil
	}

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: loginPath}
	payload := bytes.NewReader(jsonData)

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
	req.Header.Set("local", "true")
	req.Header.Set("Authorization", dm.Session.UUID)

	return req
}
