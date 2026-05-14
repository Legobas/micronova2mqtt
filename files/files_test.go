package files

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetSessionToken tests setting a single access token
func TestSetSessionToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	newToken := "test_access_token_12345"

	err := dm.SetSessionToken(newToken)
	if err != nil {
		t.Fatalf("SetSessionToken failed: %v", err)
	}

	if dm.Session.Token != newToken {
		t.Errorf("Expected token %q, got %q", newToken, dm.Session.Token)
	}

	// Verify persistence by reloading
	dm.loadSession()
	if dm.Session.Token != newToken {
		t.Errorf("Token not persisted: expected %q, got %q", newToken, dm.Session.Token)
	}
}

// TestSetSessionToken_EmptyToken tests that empty tokens are rejected
func TestSetSessionToken_EmptyToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	err := dm.SetSessionToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}

	expectedMsg := "token cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

// TestSetSessionRefreshToken tests setting a single refresh token
func TestSetSessionRefreshToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	newRefreshToken := "test_refresh_token_67890"

	err := dm.SetSessionRefreshToken(newRefreshToken)
	if err != nil {
		t.Fatalf("SetSessionRefreshToken failed: %v", err)
	}

	if dm.Session.RefreshToken != newRefreshToken {
		t.Errorf("Expected refresh token %q, got %q", newRefreshToken, dm.Session.RefreshToken)
	}

	// Verify persistence by reloading
	dm.loadSession()
	if dm.Session.RefreshToken != newRefreshToken {
		t.Errorf("Refresh token not persisted: expected %q, got %q", newRefreshToken, dm.Session.RefreshToken)
	}
}

// TestSetSessionRefreshToken_EmptyToken tests that empty refresh tokens are rejected
func TestSetSessionRefreshToken_EmptyToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	err := dm.SetSessionRefreshToken("")
	if err == nil {
		t.Error("Expected error for empty refresh token, got nil")
	}

	expectedMsg := "refresh token cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

// TestSetSessionTokens tests setting both tokens atomically
func TestSetSessionTokens(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	newToken := "test_access_token_12345"
	newRefreshToken := "test_refresh_token_67890"

	err := dm.SetSessionTokens(newToken, newRefreshToken)
	if err != nil {
		t.Fatalf("SetSessionTokens failed: %v", err)
	}

	if dm.Session.Token != newToken {
		t.Errorf("Expected token %q, got %q", newToken, dm.Session.Token)
	}

	if dm.Session.RefreshToken != newRefreshToken {
		t.Errorf("Expected refresh token %q, got %q", newRefreshToken, dm.Session.RefreshToken)
	}

	// Verify persistence by reloading
	dm.loadSession()
	if dm.Session.Token != newToken {
		t.Errorf("Token not persisted: expected %q, got %q", newToken, dm.Session.Token)
	}
	if dm.Session.RefreshToken != newRefreshToken {
		t.Errorf("Refresh token not persisted: expected %q, got %q", newRefreshToken, dm.Session.RefreshToken)
	}
}

// TestSetSessionTokens_EmptyAccessToken tests that empty access tokens are rejected
func TestSetSessionTokens_EmptyAccessToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	err := dm.SetSessionTokens("", "valid_refresh_token")
	if err == nil {
		t.Error("Expected error for empty access token, got nil")
	}

	expectedMsg := "token cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

// TestSetSessionTokens_EmptyRefreshToken tests that empty refresh tokens are rejected
func TestSetSessionTokens_EmptyRefreshToken(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	err := dm.SetSessionTokens("valid_token", "")
	if err == nil {
		t.Error("Expected error for empty refresh token, got nil")
	}

	expectedMsg := "refresh token cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
	}
}

// TestSetSessionToken_PreservesOtherFields tests that setting a token preserves other session fields
func TestSetSessionToken_PreservesOtherFields(t *testing.T) {
	dm := setupTestDataManager(t)
	defer cleanupTestDataManager(t, dm)

	// Set initial session data
	dm.Session.UUID = "test-uuid-123"
	dm.Session.ProductId = "test-product-id"
	dm.Session.DeviceId = "test-device-id"
	dm.Session.RefreshToken = "original-refresh-token"

	newToken := "new-access-token"
	err := dm.SetSessionToken(newToken)
	if err != nil {
		t.Fatalf("SetSessionToken failed: %v", err)
	}

	// Verify other fields are preserved
	if dm.Session.UUID != "test-uuid-123" {
		t.Errorf("UUID was modified: expected test-uuid-123, got %q", dm.Session.UUID)
	}
	if dm.Session.ProductId != "test-product-id" {
		t.Errorf("ProductId was modified: expected test-product-id, got %q", dm.Session.ProductId)
	}
	if dm.Session.DeviceId != "test-device-id" {
		t.Errorf("DeviceId was modified: expected test-device-id, got %q", dm.Session.DeviceId)
	}
	if dm.Session.RefreshToken != "original-refresh-token" {
		t.Errorf("RefreshToken was modified: expected original-refresh-token, got %q", dm.Session.RefreshToken)
	}
	if dm.Session.Token != newToken {
		t.Errorf("Token not updated: expected %q, got %q", newToken, dm.Session.Token)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// setupTestDataManager creates a temporary DataManager for testing
func setupTestDataManager(t *testing.T) *DataManager {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "micronova-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a minimal test config file
	configContent := `mqtt:
  url: "mqtt://localhost:1883"
  username: "test"
  password: "test"
  qos: 1
  retain: false

micronova:
  brand: "testbrand"
  email: "test@example.com"
  password: "testpass"
  power:
    on: "ON"
    off: "OFF"
  reg_keys: []
`

	configPath := filepath.Join(tmpDir, "micronova2mqtt.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	dm := &DataManager{
		Config:      Config{},
		Session:     Session{},
		configPath:  configPath,
		dataDir:     tmpDir,
		sessionPath: filepath.Join(tmpDir, "session.yml"),
	}

	return dm
}

// cleanupTestDataManager removes temporary test files
func cleanupTestDataManager(t *testing.T, dm *DataManager) {
	if err := os.RemoveAll(dm.dataDir); err != nil {
		t.Logf("Failed to cleanup temp directory: %v", err)
	}
}
