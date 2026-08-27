package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
)

const babHeader = "# bab-managed hook -- do not edit"

func FindHooksDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository (or git is not installed)")
	}
	gitDir := strings.TrimSpace(string(out))

	hooksPath, err := exec.Command("git", "config", "core.hooksPath").Output()
	if err == nil && strings.TrimSpace(string(hooksPath)) != "" {
		p := strings.TrimSpace(string(hooksPath))
		if !filepath.IsAbs(p) {
			p = filepath.Join(filepath.Dir(gitDir), p)
		}
		return p, nil
	}

	return filepath.Join(gitDir, "hooks"), nil
}

func GenerateScript(hookName string, hook *babfile.HookDef) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString(babHeader + "\n")
	b.WriteString(fmt.Sprintf("# hook: %s | To remove: bab hooks uninstall\n", hookName))
	b.WriteString("set -e\n")

	maxPos := babfile.MaxHookArgPosition(hookName)
	if maxPos > 0 {
		args := make([]string, maxPos)
		for i := range args {
			args[i] = fmt.Sprintf(`"$%d"`, i+1)
		}
		b.WriteString(fmt.Sprintf("bab hooks run %s %s\n", hookName, strings.Join(args, " ")))
	} else {
		b.WriteString(fmt.Sprintf("bab hooks run %s\n", hookName))
	}

	return b.String()
}

func isBabManaged(content string) bool {
	return strings.Contains(content, babHeader)
}

type HookStatus struct {
	Name      string
	Installed bool
	Managed   bool
	Conflict  bool
}

func Install(hooks babfile.HookMap, hooksDir string, force bool) error {
	if err := os.MkdirAll(hooksDir, 0o750); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	for name, hook := range hooks {
		hookPath := filepath.Join(hooksDir, name)
		script := GenerateScript(name, hook)

		existing, err := os.ReadFile(hookPath)
		if err == nil {
			switch {
			case isBabManaged(string(existing)):
			case !force:
				return fmt.Errorf("hook %q already exists at %s and is not managed by bab (use --force to overwrite)", name, hookPath)
			default:
				backupPath := hookPath + ".bak"
				if err := os.WriteFile(backupPath, existing, 0o600); err != nil {
					return fmt.Errorf("backing up hook %q: %w", name, err)
				}
			}
		}

		if err := os.WriteFile(hookPath, []byte(script), 0o700); err != nil { //nolint:gosec
			return fmt.Errorf("writing hook %q: %w", name, err)
		}
	}

	return nil
}

func Uninstall(hooks babfile.HookMap, hooksDir string) error {
	for name := range hooks {
		hookPath := filepath.Join(hooksDir, name)

		existing, err := os.ReadFile(hookPath)
		if err != nil {
			continue
		}

		if !isBabManaged(string(existing)) {
			continue
		}

		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("removing hook %q: %w", name, err)
		}

		backupPath := hookPath + ".bak"
		if backup, err := os.ReadFile(backupPath); err == nil {
			if err := os.WriteFile(hookPath, backup, 0o700); err != nil { //nolint:gosec
				return fmt.Errorf("restoring backup for hook %q: %w", name, err)
			}
			_ = os.Remove(backupPath)
		}
	}

	return nil
}

func Status(hooks babfile.HookMap, hooksDir string) []HookStatus {
	statuses := make([]HookStatus, 0, len(hooks))

	for name := range hooks {
		hookPath := filepath.Join(hooksDir, name)
		hs := HookStatus{Name: name}

		existing, err := os.ReadFile(hookPath)
		if err != nil {
			statuses = append(statuses, hs)
			continue
		}

		hs.Installed = true
		if isBabManaged(string(existing)) {
			hs.Managed = true
		} else {
			hs.Conflict = true
		}

		statuses = append(statuses, hs)
	}

	return statuses
}
