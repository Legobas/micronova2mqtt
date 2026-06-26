package micronova

import (
	"context"
	"encoding/json"
	"io"
	"micronova2mqtt/files"
	"net/http"
	"testing"
)

const (
	r_offset = 123
	r_mask   = 456
)

func initReadDeviceTest() {
	apiDomain = "micronova.com"
	customerCode = "182469"

	dm = &files.DataManager{}

	dm.Session.Token = "88262ea1-2934-4bdc-869e-af904a8dd5fc"
	dm.Session.ProductId = "1200"
	dm.Session.DeviceId = "56789"

	parameters = []parameter{
		{regKey: "param_get", offset: r_offset, mask: r_mask},
	}
}

// TestBuildReadDeviceRequest verifies the request method, URL, headers and JSON body.
func TestBuildReadDeviceRequest(t *testing.T) {
	initReadDeviceTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildReadDeviceRequest(ctx)
	if req == nil {
		t.Fatal("buildReadDeviceRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + readDevicePath
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
	var rdReq readDeviceReq
	if err := json.Unmarshal(bodyBytes, &rdReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if rdReq.ProductId != dm.Session.ProductId {
		t.Fatalf("values: expected %v, got %v", dm.Session.ProductId, rdReq.ProductId)
	}
	if rdReq.DeviceId != dm.Session.DeviceId {
		t.Fatalf("values: expected %v, got %v", dm.Session.DeviceId, rdReq.DeviceId)
	}
	if rdReq.Items[0] != r_offset {
		t.Fatalf("items: expected %v, got %v", r_offset, rdReq.Items)
	}
	if rdReq.Masks[0] != r_mask {
		t.Fatalf("masks: expected %v, got %v", r_mask, rdReq.Masks)
	}
}
