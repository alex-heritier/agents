package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func buildCLI(t *testing.T) string {
	t.Helper()

	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", "agents-test", "./src")
	cmd.Dir = projectRoot
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v\n%s", err, out.String())
	}

	return filepath.Join(projectRoot, "agents-test")
}

func cleanupBinary(binPath string) {
	os.Remove(binPath)
}

func createTestDir(t *testing.T, name string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	return dir
}

func createAgentsFile(t *testing.T, dir, content string) {
	t.Helper()

	path := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create AGENTS.md: %v", err)
	}
}

func runCLI(t *testing.T, binPath, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
