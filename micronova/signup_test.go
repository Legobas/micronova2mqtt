package micronova

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestBuildSignupRequest verifies the request method, URL, headers and JSON body.
func TestBuildSignupRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildSignupRequest(ctx)
	if req == nil {
		t.Fatalf("buildSignupRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + signupPath
	if req.URL.String() != expectedURL {
		t.Fatalf("expected URL %s, got %s", expectedURL, req.URL.String())
	}

	// Headers
	if got := req.Header.Get("Accept"); got != acceptHeader {
		t.Fatalf("Accept header: expected %s, got %s", acceptHeader, got)
	}
	if got := req.Header.Get("Content-Type"); got != contentTypeHeader {
		t.Fatalf("Content-Type header: expected %s, got %s", contentTypeHeader, got)
	}
	if got := req.Header.Get("id_brand"); got != brandIdHeader {
		t.Fatalf("id_brand header: expected %s, got %s", brandIdHeader, got)
	}
	if got := req.Header.Get("customer_code"); got != customerCode {
		t.Fatalf("customer_code header: expected %s, got %s", customerCode, got)
	}

	// Body: read and decode JSON
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll(body) error: %v", err)
	}
	var sReq signupReq
	if err := json.Unmarshal(bodyBytes, &sReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}

	// Verify constant defaults
	if sReq.PhoneType != PhoneType {
		t.Fatalf("PhoneType = %q, want %q", sReq.PhoneType, PhoneType)
	}
	if sReq.PhoneVersion != PhoneVersion {
		t.Fatalf("PhoneVersion = %q, want %q", sReq.PhoneVersion, PhoneVersion)
	}
	if sReq.Language != PhoneLanguage {
		t.Fatalf("Language = %q, want %q", sReq.Language, PhoneLanguage)
	}
	if sReq.PushNotificationActive != false {
		t.Fatalf("PushNotificationActive = %v, want false", sReq.PushNotificationActive)
	}

	// buildSignupRequest uses dm.CreateUUID() for several fields; just assert non-empty.
	if sReq.PhoneId == "" {
		t.Fatalf("PhoneId empty (UUID not set)")
	}
	if sReq.AppId == "" {
		t.Fatalf("AppId empty (UUID not set)")
	}
	if sReq.PushNotificationToken == "" {
		t.Fatalf("PushNotificationToken empty (UUID not set)")
	}
}
