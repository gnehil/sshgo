package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "del"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDelete(args[0])
	},
}

func runDelete(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if !cfg.RemoveProfile(name) {
		return fmt.Errorf("profile %q not found", name)
	}
	cfgPath, _ := config.DefaultConfigPath()
	if err := config.SaveConfig(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Printf("✓ Deleted %s\n", name)
	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}