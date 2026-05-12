package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/credential"
	"github.com/sshgo/sshgo/internal/executor"
	"github.com/sshgo/sshgo/internal/history"
	"golang.org/x/term"
)

var recentFlag bool

var connectCmd = &cobra.Command{
	Use:     "connect [name]",
	Aliases: []string{"c"},
	Short:   "Connect to a saved SSH session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if recentFlag {
			return runConnectRecent()
		}
		if len(args) == 0 {
			return cmd.Help()
		}
		return runConnect(args[0])
	},
}

func runConnect(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	profile := cfg.ResolveJumpHosts(*p)
	profile = withStoredJumpPasswords(profile, getCredentialPassword)
	histPath, _ := config.DefaultHistoryPath()
	h := history.NewTracker(histPath)
	h.Record(name)
	if err := h.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
	}

	if !executor.HasIdentityKey(profile) {
		profile, err = promptMissingJumpPasswords(profile)
		if err != nil {
			return err
		}
		password, err := credential.Get(credential.KindPassword, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: credential lookup failed: %v\n", err)
		}
		if password != "" {
			return executor.ExecWithPassword(password, profile)
		}
		fmt.Printf("Password for %s (%s@%s): ", name, profile.User, profile.Host)
		pwd, err := readPassword()
		if err != nil {
			return err
		}
		if len(pwd) > 0 {
			fmt.Printf("Save password to keychain? [y/N]: ")
			var ans string
			fmt.Scanln(&ans)
			if ans == "y" || ans == "Y" {
				if err := credential.Set(credential.KindPassword, name, string(pwd)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save password: %v\n", err)
				}
			}
			return executor.ExecWithPassword(string(pwd), profile)
		}
	}
	return executor.ExecSSH(profile)
}

func promptMissingJumpPasswords(p config.Profile) (config.Profile, error) {
	return withPromptedJumpPasswords(p, func(jump config.JumpHost) (string, error) {
		displayName := jump.Name
		if displayName == "" {
			displayName = jump.Host
		}
		fmt.Printf("Password for jump %s (%s@%s): ", displayName, jump.User, jump.Host)
		pwd, err := readPassword()
		if err != nil {
			return "", err
		}
		if len(pwd) == 0 {
			return "", fmt.Errorf("password is required for jump host %s", displayName)
		}
		if jump.Name != "" {
			fmt.Printf("Save password to keychain for jump %s? [y/N]: ", displayName)
			var ans string
			fmt.Scanln(&ans)
			if ans == "y" || ans == "Y" {
				if err := credential.Set(credential.KindPassword, jump.Name, string(pwd)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save jump password: %v\n", err)
				}
			}
		}
		return string(pwd), nil
	})
}

func withPromptedJumpPasswords(p config.Profile, prompt func(config.JumpHost) (string, error)) (config.Profile, error) {
	for i := range p.JumpHosts {
		jump := &p.JumpHosts[i]
		if jump.Password != "" || jump.IdentityFile != "" {
			continue
		}
		password, err := prompt(*jump)
		if err != nil {
			return p, err
		}
		jump.Password = password
	}
	return p, nil
}

func withStoredJumpPasswords(p config.Profile, lookup func(string) (string, error)) config.Profile {
	for i := range p.JumpHosts {
		jump := &p.JumpHosts[i]
		if jump.Password != "" || jump.Name == "" {
			continue
		}
		password, err := lookup(jump.Name)
		if err == nil && password != "" {
			jump.Password = password
		}
	}
	return p
}

func getCredentialPassword(name string) (string, error) {
	return credential.Get(credential.KindPassword, name)
}

func runConnectRecent() error {
	histPath, err := config.DefaultHistoryPath()
	if err != nil {
		return err
	}
	h := history.NewTracker(histPath)
	recent := h.Recent(10)
	if len(recent) == 0 {
		return fmt.Errorf("no connection history")
	}
	fmt.Println("Recent connections:")
	for i, r := range recent {
		fmt.Printf("  %d) %s (last: %s, count: %d)\n", i+1, r.Name, r.LastConnected.Format("2006-01-02 15:04"), r.ConnectCount)
	}
	fmt.Print("\nSelect a number: ")
	var n int
	if _, err := fmt.Scanln(&n); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	if n < 1 || n > len(recent) {
		return fmt.Errorf("invalid selection")
	}
	return runConnect(recent[n-1].Name)
}

func readPassword() ([]byte, error) {
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return pw, err
}

func init() {
	connectCmd.Flags().BoolVar(&recentFlag, "recent", false, "Show recent connections")
	rootCmd.AddCommand(connectCmd)
}
