package micronova

import (
	"context"
	"encoding/json"
	"io"
	"micronova2mqtt/files"
	"net/http"
	"testing"
)

func initReadDeviceBufferTest() {
	apiDomain = "micronova.com"
	customerCode = "777777"

	dm = &files.DataManager{}

	dm.Session.Token = "76132ea1-1932-3bdc-769e-af904a8dd5fc"
	dm.Session.ProductId = "9999"
	dm.Session.DeviceId = "888888"
}

// TestBuildReadDeviceBufferRequest verifies the request method, URL, headers and JSON body.
func TestBuildReadDeviceBufferRequest(t *testing.T) {
	initReadDeviceBufferTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildReadDeviceBufferRequest(ctx)
	if req == nil {
		t.Fatal("buildReadDeviceBufferRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + readDeviceBufferPath
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
	var rdbReq readDeviceBufferReq
	if err := json.Unmarshal(bodyBytes, &rdbReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if rdbReq.ProductId != dm.Session.ProductId {
		t.Fatalf("values: expected %v, got %v", dm.Session.ProductId, rdbReq.ProductId)
	}
	if rdbReq.DeviceId != dm.Session.DeviceId {
		t.Fatalf("values: expected %v, got %v", dm.Session.DeviceId, rdbReq.DeviceId)
	}
	if rdbReq.BufferId != 1 {
		t.Fatalf("items: expected %v, got %v", 1, rdbReq.BufferId)
	}
}
