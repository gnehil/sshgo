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
	histPath, _ := config.DefaultHistoryPath()
	h := history.NewTracker(histPath)
	h.Record(name)
	_ = h.Save()

	if !executor.HasIdentityKey(*p) {
		password, _ := credential.Get(credential.KindPassword, name)
		if password != "" {
			return executor.ExecWithPassword(password, *p)
		}
		fmt.Printf("Password for %s (%s@%s): ", name, p.User, p.Host)
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
			return executor.ExecWithPassword(string(pwd), *p)
		}
	}
	return executor.ExecSSH(*p)
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
	fmt.Scanln(&n)
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
