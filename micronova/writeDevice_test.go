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
	w_item  = 123
	w_value = 456789
	w_mask  = 255
)

func initWriteDeviceTest() {
	apiDomain = "micronova.com"
	customerCode = "824583"

	dm = &files.DataManager{}
	dm.Session.Token = "89262ea0-2934-4bdc-769e-af904a8dd5fc"
	dm.Session.ProductId = "1078"
	dm.Session.DeviceId = "67762"
}

// TestBuildWriteDeviceRequest verifies the request method, URL, headers and JSON body.
func TestBuildWriteDeviceRequest(t *testing.T) {
	initWriteDeviceTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildWriteDeviceRequest(ctx, w_item, w_value, w_mask)
	if req == nil {
		t.Fatal("buildWriteDeviceRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + writeDevicePath
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
	var wdReq writeDeviceReq
	if err := json.Unmarshal(bodyBytes, &wdReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if wdReq.ProductId != dm.Session.ProductId {
		t.Fatalf("values: expected %v, got %v", dm.Session.ProductId, wdReq.ProductId)
	}
	if wdReq.DeviceId != dm.Session.DeviceId {
		t.Fatalf("values: expected %v, got %v", dm.Session.DeviceId, wdReq.DeviceId)
	}
	if wdReq.Items[0] != w_item {
		t.Fatalf("items: expected %v, got %v", w_item, wdReq.Items)
	}
	if wdReq.Values[0] != w_value {
		t.Fatalf("values: expected %v, got %v", w_value, wdReq.Values)
	}
	if wdReq.Masks[0] != w_mask {
		t.Fatalf("masks: expected %v, got %v", w_mask, wdReq.Masks)
	}
}
