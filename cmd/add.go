package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var (
	addHost         string
	addPort         int
	addUser         string
	addGroup        string
	addIdentityFile string
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new SSH connection profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args[0])
	},
}

func runAdd(name string) error {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	return runAddWithConfig(cfgPath, name, addHost, addPort, addUser, addGroup, addIdentityFile)
}

func runAddWithConfig(cfgPath, name, host string, port int, user, group, identityFile string) error {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}
	if cfg.FindProfile(name) != nil {
		return fmt.Errorf("profile %q already exists", name)
	}
	p := config.Profile{
		Name:         name,
		Host:         host,
		Port:         port,
		User:         user,
		Group:        group,
		IdentityFile: identityFile,
	}
	if err := p.Validate(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	cfg.AddProfile(p)
	if err := config.SaveConfig(cfgPath, cfg); err != nil {
		return err
	}
	fmt.Printf("[OK] Added %s (%s@%s:%d)\n", name, user, host, port)
	return nil
}

func init() {
	addCmd.Flags().StringVar(&addHost, "host", "", "SSH host (required)")
	addCmd.Flags().IntVarP(&addPort, "port", "p", 22, "SSH port")
	addCmd.Flags().StringVarP(&addUser, "user", "u", "", "SSH user (required)")
	addCmd.Flags().StringVar(&addGroup, "group", "", "Group name")
	addCmd.Flags().StringVarP(&addIdentityFile, "identity-file", "i", "", "Path to identity (private key) file")
	addCmd.MarkFlagRequired("host")
	addCmd.MarkFlagRequired("user")
	rootCmd.AddCommand(addCmd)
}
