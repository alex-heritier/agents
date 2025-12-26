package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListCommand(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "list_test")

	subDir1 := filepath.Join(testDir, "subdir1")
	subDir2 := filepath.Join(testDir, "subdir2")
	os.MkdirAll(subDir1, 0755)
	os.MkdirAll(subDir2, 0755)

	createAgentsFile(t, testDir, "Root agents file")
	createAgentsFile(t, subDir1, "Subdir1 agents file")
	createAgentsFile(t, subDir2, "Subdir2 agents file")

	output, err := runCLI(t, bin, testDir, "rule", "list")
	if err != nil {
		t.Errorf("rule list command failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "AGENTS.md") {
		t.Errorf("list output missing AGENTS.md: %s", output)
	}
}

func TestListWithVerbose(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "verbose_test")
	createAgentsFile(t, testDir, "Test content")

	output, err := runCLI(t, bin, testDir, "rule", "list", "--verbose")
	if err != nil {
		t.Errorf("rule list --verbose failed: %v\n%s", err, output)
	}
}

func TestListWithGlobalFlag(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "global_test")

	output, err := runCLI(t, bin, testDir, "rule", "list", "--global")
	if err != nil {
		t.Errorf("rule list --global failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Directory") {
		t.Error("expected 'Directory' header in output")
	}
}

func TestSyncDryRun(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "dryrun_test")
	createAgentsFile(t, testDir, "Test content")

	output, err := runCLI(t, bin, testDir, "rule", "sync", "--claude", "--dry-run")
	if err != nil {
		t.Errorf("rule sync --dry-run failed: %v\n%s", err, output)
	}

	claudePath := filepath.Join(testDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); !os.IsNotExist(err) {
		t.Error("CLAUDE.md was created in dry-run mode")
	}
}

func TestSyncCreatesSymlink(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "sync_test")
	createAgentsFile(t, testDir, "Test content")

	_, err := runCLI(t, bin, testDir, "rule", "sync", "--claude")
	if err != nil {
		t.Errorf("rule sync command failed: %v", err)
	}

	claudePath := filepath.Join(testDir, "CLAUDE.md")
	info, err := os.Lstat(claudePath)
	if err != nil {
		t.Errorf("CLAUDE.md not created: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("CLAUDE.md is not a symlink")
	}
}

func TestSyncVerbose(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "verbose_sync_test")
	createAgentsFile(t, testDir, "Test content")

	output, err := runCLI(t, bin, testDir, "rule", "sync", "--claude", "--verbose")
	if err != nil {
		t.Errorf("rule sync --verbose failed: %v\n%s", err, output)
	}
}

func TestRmCommand(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "rm_test")
	createAgentsFile(t, testDir, "Test content")

	_, err := runCLI(t, bin, testDir, "rule", "sync", "--claude")
	if err != nil {
		t.Errorf("setup sync failed: %v", err)
	}

	claudePath := filepath.Join(testDir, "CLAUDE.md")
	lstatInfo, lstatErr := os.Lstat(claudePath)
	if lstatErr != nil {
		if os.IsNotExist(lstatErr) {
			t.Error("CLAUDE.md not created for rm test")
		} else {
			t.Errorf("error checking CLAUDE.md: %v", lstatErr)
		}
	} else {
		isSymlink := lstatInfo.Mode()&os.ModeSymlink != 0
		t.Logf("CLAUDE.md before rm: mode=%v, isSymlink=%v", lstatInfo.Mode(), isSymlink)
		if !isSymlink {
			t.Errorf("CLAUDE.md is not a symlink, it's a regular file")
		}
	}

	rmOutput, rmErr := runCLI(t, bin, testDir, "rule", "rm", "--claude")
	t.Logf("rm output: %s", rmOutput)
	if rmErr != nil {
		t.Errorf("rule rm command failed: %v\n%s", rmErr, rmOutput)
	}

	_, statErr := os.Stat(claudePath)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			t.Errorf("unexpected error checking CLAUDE.md: %v", statErr)
		}
	} else {
		t.Error("CLAUDE.md still exists after rm")
	}
}

func TestRmDryRun(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "rm_dryrun_test")
	createAgentsFile(t, testDir, "Test content")

	_, err := runCLI(t, bin, testDir, "rule", "sync", "--claude")
	if err != nil {
		t.Errorf("setup sync failed: %v", err)
	}

	runCLI(t, bin, testDir, "rule", "rm", "--claude", "--dry-run")

	claudePath := filepath.Join(testDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md was removed in dry-run mode")
	}
}

func TestMultipleAgents(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "multi_agent_test")
	createAgentsFile(t, testDir, "Test content")

	_, err := runCLI(t, bin, testDir, "rule", "sync", "--claude", "--cursor")
	if err != nil {
		t.Errorf("rule sync with multiple agents failed: %v", err)
	}

	claudePath := filepath.Join(testDir, "CLAUDE.md")
	cursorPath := filepath.Join(testDir, ".cursor", "rules", "agents.md")

	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md not created")
	}

	if _, err := os.Stat(cursorPath); os.IsNotExist(err) {
		t.Error(".cursor/rules/agents.md not created")
	}
}

func TestRuleModuleUnknownFlag(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "rule_unknown_flag")
	createAgentsFile(t, testDir, "Test content")

	testCases := []struct {
		command string
		flag    string
	}{
		{"list", "--unknown-flag"},
		{"sync", "--bad-flag"},
		{"rm", "--invalid"},
	}

	for _, tc := range testCases {
		output, err := runCLI(t, bin, testDir, "rule", tc.command, tc.flag)
		if err == nil {
			t.Errorf("expected error for rule %s with %s", tc.command, tc.flag)
		}
		if !strings.Contains(output, "Unknown flags") {
			t.Errorf("expected 'Unknown flags' in output for rule %s %s", tc.command, tc.flag)
		}
	}
}
