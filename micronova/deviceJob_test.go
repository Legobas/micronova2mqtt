package micronova

import (
	"context"
	"micronova2mqtt/files"
	"net/http"
	"testing"
)

const (
	requestId = "12345678"
)

func initDeviceJobTest() {
	apiDomain = "micronova.com"
	customerCode = "888888"

	dm = &files.DataManager{}

	dm.Session.Token = "88262ea1-2934-4bdc-869e-af904a8dd5fc"
	dm.Session.ProductId = "1"
	dm.Session.DeviceId = "2"
}

// TestBuildDeviceJobRequest verifies the request method, URL, headers and JSON body.
func TestBuildDeviceJobRequest(t *testing.T) {
	initDeviceJobTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceJobRequest(ctx, requestId)
	if req == nil {
		t.Fatal("buildDeviceJobRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodGet {
		t.Fatalf("expected method GET, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + deviceJobPath + requestId
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
	if got := req.Header.Get("local"); got != "false" {
		t.Fatalf("local header: expected false, got %s", got)
	}
	if got := req.Header.Get("Authorization"); got != dm.Session.Token {
		t.Fatalf("Authorization header: expected %s, got %s", dm.Session.Token, got)
	}
}
