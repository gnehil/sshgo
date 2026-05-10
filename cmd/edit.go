package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a profile (opens config in editor)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return openEditorDefault()
		}
		return editProfile(args[0])
	},
}

func editProfile(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if cfg.FindProfile(name) == nil {
		return fmt.Errorf("配置 %q 不存在", name)
	}
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	return openEditor(cfgPath)
}

func openEditorDefault() error {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	return openEditor(cfgPath)
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	command := exec.Command(editor, path)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func init() {
	rootCmd.AddCommand(editCmd)
}