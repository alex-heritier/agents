package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md not created for rm test")
	}

	_, err = runCLI(t, bin, testDir, "rule", "rm", "--claude")
	if err != nil {
		t.Errorf("rule rm command failed: %v", err)
	}

	if _, err := os.Stat(claudePath); !os.IsNotExist(err) {
		t.Error("CLAUDE.md was not removed")
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

func TestListCommands(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "list_commands_test")

	output, err := runCLI(t, bin, testDir, "command", "list")
	if err != nil {
		t.Errorf("command list failed: %v\n%s", err, output)
	}
}

func TestListSkills(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "list_skills_test")

	output, err := runCLI(t, bin, testDir, "skill", "list")
	if err != nil {
		t.Errorf("skill list failed: %v\n%s", err, output)
	}
}

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

	modules := []string{"rule", "command", "skill"}

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

func TestCommandModuleUnknownFlag(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "command_unknown_flag")
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
		output, err := runCLI(t, bin, testDir, "command", tc.command, tc.flag)
		if err == nil {
			t.Errorf("expected error for command %s with %s", tc.command, tc.flag)
		}
		if !strings.Contains(output, "Unknown flags") {
			t.Errorf("expected 'Unknown flags' in output for command %s %s", tc.command, tc.flag)
		}
	}
}

func TestSkillModuleUnknownFlag(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "skill_unknown_flag")

	testCases := []struct {
		command string
		flag    string
	}{
		{"list", "--unknown-flag"},
		{"sync", "--bad-flag"},
	}

	for _, tc := range testCases {
		output, err := runCLI(t, bin, testDir, "skill", tc.command, tc.flag)
		if err == nil {
			t.Errorf("expected error for skill %s with %s", tc.command, tc.flag)
		}
		if !strings.Contains(output, "Unknown flags") {
			t.Errorf("expected 'Unknown flags' in output for skill %s %s", tc.command, tc.flag)
		}
	}
}

func TestCommandSyncNestedSymlinks(t *testing.T) {
	bin := buildCLI(t)
	defer cleanupBinary(bin)

	testDir := createTestDir(t, "nested_sync_test")

	createAgentsFile(t, testDir, "Test content")

	commandsPath := filepath.Join(testDir, "COMMANDS.md")
	if err := os.WriteFile(commandsPath, []byte("# Commands"), 0644); err != nil {
		t.Fatalf("Failed to create COMMANDS.md: %v", err)
	}

	_, err := runCLI(t, bin, testDir, "command", "sync", "--cursor")
	if err != nil {
		t.Errorf("command sync failed: %v", err)
	}

	cursorCommandsPath := filepath.Join(testDir, ".cursor", "commands", "commands.md")
	info, err := os.Lstat(cursorCommandsPath)
	if err != nil {
		t.Errorf(".cursor/commands/commands.md not created: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error(".cursor/commands/commands.md is not a symlink")
	}

	target, err := os.Readlink(cursorCommandsPath)
	if err != nil {
		t.Errorf("Failed to read symlink: %v", err)
	}

	expectedTarget := "../../COMMANDS.md"
	if target != expectedTarget {
		t.Errorf("symlink target is %s, expected %s", target, expectedTarget)
	}
}
