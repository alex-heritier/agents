package e2e

import (
	"strings"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	output, err := runCLI(t, bin, ".", "help")
	if err != nil {
		t.Errorf("help command failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Usage:") {
		t.Error("help output missing Usage")
	}
}

func TestUnknownCommand(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	output, err := runCLI(t, bin, ".", "unknown-module")
	if err == nil {
		t.Error("expected error for unknown module")
	}

	if !strings.Contains(output, "Unknown module") {
		t.Error("expected 'Unknown module' in output")
	}
}

func TestModuleOnlyShowsHelp(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	modules := []string{"rule"}

	for _, module := range modules {
		output, err := runCLI(t, bin, ".", module)
		if err != nil {
			t.Errorf("expected no error for module-only invocation '%s', got: %v", module, err)
		}

		if !strings.Contains(output, "Usage:") {
			t.Errorf("expected 'Usage:' in output for module '%s'", module)
		}

		if !strings.Contains(output, module) {
			t.Errorf("expected module name '%s' in help output", module)
		}
	}
}

func TestUnknownFlag(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "unknown_flag_test")
	createAgentsFile(t, testDir, "Test content")

	output, err := runCLI(t, bin, testDir, "rule", "list", "--unknown-flag")
	if err == nil {
		t.Error("expected error for unknown flag")
	}

	if !strings.Contains(output, "Unknown flags") {
		t.Error("expected 'Unknown flags' in output")
	}
}
