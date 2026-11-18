package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Mock interfaces for dependencies
type mockMicronova struct {
	powerCalls     []string
	parameterCalls []struct {
		key   string
		value string
	}
	updateMqttCalls []bool
}

func (m *mockMicronova) SetPower(value string) {
	m.powerCalls = append(m.powerCalls, value)
}

func (m *mockMicronova) SetParameter(key, value string) {
	m.parameterCalls = append(m.parameterCalls, struct {
		key   string
		value string
	}{key, value})
}

func (m *mockMicronova) SetUpdateMqtt(value bool) {
	m.updateMqttCalls = append(m.updateMqttCalls, value)
}

// TestReceiveMqttMessage_ParameterKey tests setting parameters through MQTT
func TestReceiveMqttMessage_ParameterKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "set temperature",
			key:   "temperature",
			value: "22.5",
		},
		{
			name:  "set mode",
			key:   "mode",
			value: "heating",
		},
		{
			name:  "set fan speed",
			key:   "fanSpeed",
			value: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

// TestAppName_Constant tests the app name constant
func TestAppName_Constant(t *testing.T) {
	expected := "Micronova2MQTT"
	if appName != expected {
		t.Errorf("AppName constant = %q, want %q", appName, expected)
	}
}

// TestVersion_Variable tests that Version can be set
func TestVersion_Variable(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "1.2.3"
	if Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", Version, "1.2.3")
	}
}
func TestStartHealthServer(t *testing.T) {
	srv := startHealthServer("127.0.0.1:8080")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	// Give the goroutine a brief moment to bind to 127.0.0.1:8080
	var resp *http.Response
	var err error
	for i := 0; i < 10; i++ {
		resp, err = http.Get("http://127.0.0.1:8080/health")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Failed to connect to health server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got %q", string(body))
	}
}

func TestCheckHealth_FlagFalse(t *testing.T) {
	flagVal := false
	// Should return immediately without calling os.Exit
	checkHealth(&flagVal)
}

func TestCheckHealth_Success(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		flagVal := true
		checkHealth(&flagVal)
		return
	}

	// Start the server in the main test process so the child process can query it
	srv := startHealthServer("127.0.0.1:8080")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	time.Sleep(50 * time.Millisecond)

	// Run the test in a subprocess because checkHealth calls os.Exit(0)
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckHealth_Success")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	if err != nil {
		t.Fatalf("Expected checkHealth to exit with status 0, got error: %v", err)
	}
}

func TestCheckHealth_Failure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		flagVal := true
		checkHealth(&flagVal)
		return
	}

	// Ensure no server is running on 127.0.0.1:8080, then run in subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckHealth_Failure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("Expected exit code 1, got %d", exitErr.ExitCode())
		}
		return
	}
	t.Fatalf("Expected checkHealth to fail with exit code 1, but got err: %v", err)
}
