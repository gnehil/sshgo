package cmd

import (
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "sshgo",
	Short: "SSH session manager CLI",
	Long:  "sshgo is a CLI tool for managing SSH sessions with support for groups, jump hosts, port forwarding, and batch execution.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return runConnect(args[0])
		}
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}