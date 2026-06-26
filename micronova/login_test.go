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
	email    = "test@example.com"
	password = "p@ssw0rd"
)

func initLoginTest() {
	apiDomain = "micronova.com"
	customerCode = "123456"

	dm = &files.DataManager{}

	dm.Config.Micronova.Email = email
	dm.Config.Micronova.Password = password
}

// TestBuildLoginRequest verifies the request method, URL, headers and JSON body.
func TestBuildLoginRequest(t *testing.T) {
	initLoginTest()
	
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := buildLoginRequest(ctx)
	if req == nil {
		t.Fatal("buildLoginRequest returned nil")
	}

	// Method and URL
	if req.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", req.Method)
	}
	expectedURL := "https://" + apiDomain + "/" + loginPath
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
	if got := req.Header.Get("local"); got != "true" {
		t.Fatalf("local header: expected true, got %s", got)
	}
	if got := req.Header.Get("Authorization"); got != dm.Session.UUID {
		t.Fatalf("Authorization header: expected %s, got %s", dm.Session.UUID, got)
	}

	// Body: read and decode JSON
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll(body) error: %v", err)
	}
	var lReq loginReq
	if err := json.Unmarshal(bodyBytes, &lReq); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if lReq.Email != dm.Config.Micronova.Email {
		t.Fatalf("email: expected %s, got %s", email, lReq.Email)
	}
	if lReq.Password != dm.Config.Micronova.Password {
		t.Fatalf("password: expected %s, got %s", password, lReq.Password)
	}
}
