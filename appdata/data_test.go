package appdata

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// setupTestEnv isolates the working directory and user home directory using temporary
// folders so tests never touch real system or user files.
func setupTestEnv(t *testing.T) (workDir string, homeDir string) {
	t.Helper()

	baseDir := t.TempDir()
	workDir = filepath.Join(baseDir, "work")
	homeDir = filepath.Join(baseDir, "home")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create temp work dir: %v", err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("failed to create temp home dir: %v", err)
	}

	// Isolate home directory across Unix and Windows
	t.Setenv("HOME", homeDir)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", homeDir)
	}

	// Isolate working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to change directory to temp work dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	return workDir, homeDir
}

func TestGetPaths(t *testing.T) {
	workDir, homeDir := setupTestEnv(t)

	// Resolve symlinks (e.g., on macOS /var -> /private/var) so os.Getwd() matches workDir
	resolvedWorkDir, err := filepath.EvalSymlinks(workDir)
	if err != nil {
		resolvedWorkDir = workDir
	}

	expected := []string{
		filepath.Join("/", dataDir, configFilename),
		filepath.Join(homeDir, dataDir, configFilename),
		filepath.Join(resolvedWorkDir, configFilename),
		filepath.Join(resolvedWorkDir, dataDir, configFilename),
	}

	got := getPaths()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("getPaths() mismatch:\ngot:  %v\nwant: %v", got, expected)
	}
}
