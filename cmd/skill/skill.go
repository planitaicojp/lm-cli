package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	lmerrors "github.com/crowdy/lm-cli/internal/errors"
	"github.com/crowdy/lm-cli/internal/prompt"
)

const (
	skillRepo = "https://github.com/planitaicojp/lm-cli-skill.git"
	skillName = "lm-cli-skill"
)

// Cmd is the parent command for skill management.
var Cmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage Claude Code skills",
}

func init() {
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(removeCmd)
}

func defaultSkillBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

func runInstall(baseDir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return &lmerrors.ValidationError{Message: "git is required to install skills"}
	}

	skillDir := filepath.Join(baseDir, skillName)
	if _, err := os.Stat(skillDir); err == nil {
		return &lmerrors.ValidationError{Message: "already installed, use 'lm skill update'"}
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("cannot create skill directory: %w", err)
	}

	cmd := exec.Command("git", "clone", skillRepo, skillDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &lmerrors.NetworkError{Err: fmt.Errorf("git clone failed: %w", err)}
	}

	fmt.Fprintln(os.Stderr, "Installed lm-cli-skill successfully.")
	return nil
}

func runUpdate(baseDir string) error {
	skillDir := filepath.Join(baseDir, skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return &lmerrors.ValidationError{Message: "not installed, use 'lm skill install'"}
	}

	gitDir := filepath.Join(skillDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return &lmerrors.ValidationError{Message: "not a git repository, remove and reinstall"}
	}

	cmd := exec.Command("git", "-C", skillDir, "pull")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &lmerrors.NetworkError{Err: fmt.Errorf("git pull failed: %w", err)}
	}

	fmt.Fprintln(os.Stderr, "Updated lm-cli-skill successfully.")
	return nil
}

func runRemove(baseDir string) error {
	skillDir := filepath.Join(baseDir, skillName)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return &lmerrors.ValidationError{Message: "not installed"}
	}

	ok, err := prompt.Confirm("Remove lm-cli-skill?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Removed lm-cli-skill successfully.")
	return nil
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install lm-cli-skill for Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		base, err := defaultSkillBase()
		if err != nil {
			return err
		}
		return runInstall(base)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update lm-cli-skill to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		base, err := defaultSkillBase()
		if err != nil {
			return err
		}
		return runUpdate(base)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove lm-cli-skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		base, err := defaultSkillBase()
		if err != nil {
			return err
		}
		return runRemove(base)
	},
}
