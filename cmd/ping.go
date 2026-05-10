package cmd

import (
	"fmt"
	"os/exec"
	"time"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/executor"
)

var pingCommand = &cobra.Command{
	Use:   "ping <name>",
	Short: "Test SSH connectivity to a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPing(args[0])
	},
}

func runPing(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	bin, args := executor.BuildSSHCommand(*p)
	args = append(args, "-o", "ConnectTimeout=5", "-o", "BatchMode=yes")
	lastArg := args[len(args)-1]
	args[len(args)-1] = lastArg + " echo 'Connection OK'"
	start := time.Now()
	command := exec.Command(bin, args...)
	err = command.Run()
	duration := time.Since(start)
	if err != nil {
		fmt.Printf("✗ %s@%s:%d - Connection failed (%v)\n", p.User, p.Host, p.Port, err)
		return nil
	}
	fmt.Printf("✓ %s@%s:%d - Connected in %v\n", p.User, p.Host, p.Port, duration.Round(time.Millisecond))
	return nil
}

func init() {
	rootCmd.AddCommand(pingCommand)
}