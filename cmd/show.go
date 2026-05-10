package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runShow(args[0])
	},
}

func runShow(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	fmt.Printf("Name:        %s\n", p.Name)
	fmt.Printf("Host:        %s\n", p.Host)
	fmt.Printf("Port:        %d\n", p.Port)
	fmt.Printf("User:        %s\n", p.User)
	if p.Group != "" {
		fmt.Printf("Group:       %s\n", p.Group)
	}
	if p.IdentityFile != "" {
		fmt.Printf("Identity:    %s\n", p.IdentityFile)
	}
	if len(p.JumpHosts) > 0 {
		fmt.Printf("Jump Hosts:\n")
		for _, j := range p.JumpHosts {
			fmt.Printf("  - %s@%s:%d\n", j.User, j.Host, j.Port)
		}
	}
	if len(p.ForwardPorts) > 0 {
		fmt.Printf("Port Forwards:\n")
		for _, f := range p.ForwardPorts {
			fmt.Printf("  - %d -> %s:%d\n", f.LocalPort, f.RemoteHost, f.RemotePort)
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(showCmd)
}