package micronova

import (
	"context"
	"encoding/json"
	"io"
	"micronova2mqtt/files"
	"net/http"
	"testing"
)

func initRefreshTest() {
	apiDomain = "micronova.com"
	customerCode = "888888"

	dm = &files.DataManager{}
	dm.Session.RefreshToken = "12345ab-1234-56dc-7890-abcdef00"
}

// TestBuildRefreshRequest verifies the request method, URL, headers and JSON body.
func TestBuildRefreshRequest(t *testing.T) {
	initRefreshTest()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildRefreshRequest(ctx)
	if req == nil {
		t.Fatal("buildRefreshRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + refreshPath
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

	var rReq map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if rReq["refresh_token"] != dm.Session.RefreshToken {
		t.Fatalf("values: expected %v, got %v", dm.Session.RefreshToken, rReq["refresh_token"])
	}
}
