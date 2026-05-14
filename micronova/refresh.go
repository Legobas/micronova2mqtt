package micronova

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

const (
	refreshPath = "refreshToken"
)

type refreshResp struct {
	Success bool   `json:"Success"`
	Text    string `json:"Text"`
	Token   string `json:"token"`
	Error   string `json:"Error"`
}

func updateToken() error {
	// Create context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildRefreshRequest(ctx)
	if req == nil {
		log.Fatal().Msg("Failed to build RefreshToken request")
	}

	log.Info().Msgf("API Call: %s", refreshPath)

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("RefreshToken HTTP Client error: %s", err.Error())
	}
	defer resp.Body.Close()

	var result refreshResp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal().Err(err).Msg("Error decoding response")
	}

	// Validate HTTP status code and response Success field
	if resp.StatusCode != http.StatusCreated || !result.Success {
		dm.SetSessionTokens("", "")
		return fmt.Errorf("RefreshToken Unsuccesful (%s) %s %s", resp.Status, result.Error, result.Text)
	}

	if result.Token == "" {
		return errors.New("RefreshToken: Token is empty")
	}

	if err := dm.SetSessionToken(result.Token); err != nil {
		return fmt.Errorf("failed to set session token: %w", err)
	}

	return nil
}

func buildRefreshRequest(ctx context.Context) *http.Request {
	jsonData := []byte(`{"refresh_token": "` + dm.Session.RefreshToken + `"}`)

	reqUrl := url.URL{Scheme: "https", Host: apiDomain, Path: refreshPath}
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

	return req
}
