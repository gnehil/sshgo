package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/sshconfig"
)

var syncDryRun bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync profiles to ~/.ssh/config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync()
	},
}

func runSync() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return sshconfig.GenerateSSHConfig(cfg, syncDryRun)
}

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(syncCmd)
}