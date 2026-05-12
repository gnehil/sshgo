package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "sshgo",
	Short:   "SSH session manager CLI",
	Long:    "sshgo is a CLI tool for managing SSH sessions with support for groups, jump hosts, port forwarding, and batch execution.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return runConnect(args[0])
		}
		return cmd.Help()
	},
}

func loadConfig() (*config.Config, error) {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	return config.LoadConfig(cfgPath)
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
}

func Execute() error {
	args := os.Args[1:]
	if cfg, err := loadConfig(); err == nil {
		if name, ok := profileShortcutName(args, cfg); ok {
			return runConnect(name)
		}
	} else if len(args) == 1 && !strings.HasPrefix(args[0], "-") && !isRootCommand(args[0]) {
		return err
	}
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func profileShortcutName(args []string, cfg *config.Config) (string, bool) {
	if len(args) != 1 || strings.HasPrefix(args[0], "-") || isRootCommand(args[0]) {
		return "", false
	}
	if cfg == nil || cfg.FindProfile(args[0]) == nil {
		return "", false
	}
	return args[0], true
}

func isRootCommand(name string) bool {
	if name == "help" {
		return true
	}
	for _, command := range rootCmd.Commands() {
		if command.Name() == name || command.HasAlias(name) {
			return true
		}
	}
	return false
}
