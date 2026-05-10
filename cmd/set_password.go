package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/credential"
	"golang.org/x/term"
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
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return err
	}
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	if err := credential.Set(credential.KindPassword, name, string(password)); err != nil {
		return fmt.Errorf("failed to store password: %w", err)
	}
	fmt.Printf("✓ Password saved for %s\n", name)
	return nil
}

func init() {
	rootCmd.AddCommand(setPasswordCmd)
}
