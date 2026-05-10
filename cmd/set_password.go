package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/credential"
)

var setPasswordCmd = &cobra.Command{
	Use:   "set-password <name>",
	Short: "Store password in OS keychain for a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetPassword(args[0])
	},
}

func runSetPassword(name string) error {
	fmt.Print("Password: ")
	password, err := readPassword()
	if err != nil {
		return err
	}
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	if err := credential.Set(credential.KindPassword, name, string(password)); err != nil {
		return fmt.Errorf("failed to store password: %w", err)
	}
	fmt.Printf("[OK] Password saved for %s\n", name)
	return nil
}

var deletePasswordCmd = &cobra.Command{
	Use:   "remove-password <name>",
	Short: "Delete saved password from keychain",
	Aliases: []string{"del-password"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeletePassword(args[0])
	},
}

func runDeletePassword(name string) error {
	if err := credential.Delete(credential.KindPassword, name); err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}
	fmt.Printf("[OK] Password removed for %s\n", name)
	return nil
}

func init() {
	rootCmd.AddCommand(setPasswordCmd)
	rootCmd.AddCommand(deletePasswordCmd)
}
