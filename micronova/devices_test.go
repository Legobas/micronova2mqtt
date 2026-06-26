package micronova

import (
	"context"
	"encoding/json"
	"io"
	"micronova2mqtt/files"
	"net/http"
	"testing"
)

func initDevicesTest() {
	apiDomain = "micronova.com"
	customerCode = "888888"

	dm = &files.DataManager{}

	dm.Session.Token = "88262ea1-2934-4bdc-869e-af904a8dd5fc"
	dm.Session.ProductId = "1"
	dm.Session.DeviceId = "2"
}

// TestBuildDeviceListRequest verifies the request method, URL, headers and JSON body.
func TestBuildDeviceListRequest(t *testing.T) {
	initDevicesTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildDeviceListRequest(ctx)
	if req == nil {
		t.Fatal("buildDeviceListRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + devicesPath
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

	// Body: read and decode JSON
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll(body) error: %v", err)
	}

	var dlReq map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &dlReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if len(dlReq) != 0 {
		t.Fatalf("JSON not empty")
	}
}
