package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
)

func TestGenerateScript(t *testing.T) {
	hook := &babfile.HookDef{
		Name: "pre-commit",
	}
	script := GenerateScript("pre-commit", hook)

	if !strings.Contains(script, "#!/bin/sh") {
		t.Error("expected shebang line")
	}
	if !strings.Contains(script, babHeader) {
		t.Error("expected bab header")
	}
	if !strings.Contains(script, "bab hooks run pre-commit\n") {
		t.Errorf("expected 'bab hooks run pre-commit' without args, got:\n%s", script)
	}
	if !strings.Contains(script, "set -e") {
		t.Error("expected set -e")
	}
}

func TestGenerateScriptWithArgs(t *testing.T) {
	hook := &babfile.HookDef{Name: "commit-msg"}
	script := GenerateScript("commit-msg", hook)

	if !strings.Contains(script, `bab hooks run commit-msg "$1"`) {
		t.Errorf("expected '$1' arg, got:\n%s", script)
	}
}

func TestGenerateScriptWithMultipleArgs(t *testing.T) {
	hook := &babfile.HookDef{Name: "pre-push"}
	script := GenerateScript("pre-push", hook)

	if !strings.Contains(script, `bab hooks run pre-push "$1" "$2"`) {
		t.Errorf("expected '$1' '$2' args, got:\n%s", script)
	}
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
	}

	if err := Install(hooks, dir, false); err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "pre-commit"))
	if err != nil {
		t.Fatalf("reading hook file: %v", err)
	}

	if !strings.Contains(string(content), babHeader) {
		t.Error("installed hook should contain bab header")
	}

	info, _ := os.Stat(filepath.Join(dir, "pre-commit"))
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("hook file should be executable")
	}
}

func TestInstallExistingNonBabHook(t *testing.T) {
	dir := t.TempDir()

	hookPath := filepath.Join(dir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho custom\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
	}

	err := Install(hooks, dir, false)
	if err == nil {
		t.Fatal("expected error when non-bab hook exists without --force")
	}
	if !strings.Contains(err.Error(), "not managed by bab") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInstallForce(t *testing.T) {
	dir := t.TempDir()

	hookPath := filepath.Join(dir, "pre-commit")
	originalContent := "#!/bin/sh\necho custom\n"
	if err := os.WriteFile(hookPath, []byte(originalContent), 0o600); err != nil {
		t.Fatal(err)
	}

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
	}

	if err := Install(hooks, dir, true); err != nil {
		t.Fatalf("Install(force=true) error: %v", err)
	}

	backup, err := os.ReadFile(hookPath + ".bak")
	if err != nil {
		t.Fatal("expected .bak backup file")
	}
	if string(backup) != originalContent {
		t.Error("backup content doesn't match original")
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), babHeader) {
		t.Error("new hook should be bab-managed")
	}
}

func TestUninstall(t *testing.T) {
	dir := t.TempDir()

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
	}

	if err := Install(hooks, dir, false); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(hooks, dir); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "pre-commit")); !os.IsNotExist(err) {
		t.Error("hook file should be removed after uninstall")
	}
}

func TestUninstallRestoresBackup(t *testing.T) {
	dir := t.TempDir()

	hookPath := filepath.Join(dir, "pre-commit")
	originalContent := "#!/bin/sh\necho custom\n"
	if err := os.WriteFile(hookPath, []byte(originalContent), 0o600); err != nil {
		t.Fatal(err)
	}

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
	}

	if err := Install(hooks, dir, true); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(hooks, dir); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal("expected hook to be restored from backup")
	}
	if string(content) != originalContent {
		t.Error("restored content doesn't match original")
	}

	if _, err := os.Stat(hookPath + ".bak"); !os.IsNotExist(err) {
		t.Error("backup file should be removed after restore")
	}
}

func TestStatus(t *testing.T) {
	dir := t.TempDir()

	hooks := babfile.HookMap{
		"pre-commit": &babfile.HookDef{Name: "pre-commit"},
		"pre-push":   &babfile.HookDef{Name: "pre-push"},
		"commit-msg": &babfile.HookDef{Name: "commit-msg"},
	}

	if err := Install(babfile.HookMap{"pre-commit": hooks["pre-commit"]}, dir, false); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "commit-msg"), []byte("#!/bin/sh\necho custom\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	statuses := Status(hooks, dir)
	statusMap := make(map[string]HookStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	if s := statusMap["pre-commit"]; !s.Installed || !s.Managed {
		t.Errorf("pre-commit should be installed and managed: %+v", s)
	}

	if s := statusMap["pre-push"]; s.Installed {
		t.Errorf("pre-push should not be installed: %+v", s)
	}

	if s := statusMap["commit-msg"]; !s.Conflict {
		t.Errorf("commit-msg should show conflict: %+v", s)
	}
}

func TestIsBabManaged(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"#!/bin/sh\n" + babHeader + "\nset -e\n", true},
		{"#!/bin/sh\necho custom\n", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isBabManaged(tt.content); got != tt.want {
			t.Errorf("isBabManaged(%q) = %v, want %v", tt.content, got, tt.want)
		}
	}
}
