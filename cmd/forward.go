package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/output"
)

var localForward string

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Manage port forwarding",
}

var forwardAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add port forwarding to a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForwardAdd(args[0])
	},
}

func runForwardAdd(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	var fp config.ForwardPort
	if localForward != "" {
		parts := strings.Split(localForward, ":")
		if len(parts) == 2 {
			localPort, err := strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("invalid local port %q: %w", parts[0], err)
			}
			fp.LocalPort = localPort
			fp.RemoteHost = "localhost"
			remotePort, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid remote port %q: %w", parts[1], err)
			}
			fp.RemotePort = remotePort
		} else if len(parts) == 3 {
			localPort, err := strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("invalid local port %q: %w", parts[0], err)
			}
			fp.LocalPort = localPort
			fp.RemoteHost = parts[1]
			remotePort, err := strconv.Atoi(parts[2])
			if err != nil {
				return fmt.Errorf("invalid remote port %q: %w", parts[2], err)
			}
			fp.RemotePort = remotePort
		} else {
			return fmt.Errorf("invalid -L format: expected local_port:remote_host:remote_port")
		}
	}
	if fp.LocalPort == 0 {
		return fmt.Errorf("use -L to specify forward rule")
	}
	p.ForwardPorts = append(p.ForwardPorts, fp)
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

var forwardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all port forwarding rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForwardList()
	},
}

func runForwardList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	var hasForwards bool
	var rows [][]string
	for _, p := range cfg.Profiles {
		for _, f := range p.ForwardPorts {
			hasForwards = true
			rows = append(rows, []string{p.Name, fmt.Sprintf("%d", f.LocalPort), fmt.Sprintf("%s:%d", f.RemoteHost, f.RemotePort)})
		}
	}
	if !hasForwards {
		fmt.Println("No port forwarding rules configured.")
		return nil
	}
	return output.PrintTable(rows, []string{"PROFILE", "LOCAL", "REMOTE"})
}

func init() {
	forwardAddCmd.Flags().StringVarP(&localForward, "local", "L", "", "Local port forwarding (local_port:remote_host:remote_port)")
	forwardCmd.AddCommand(forwardAddCmd, forwardListCmd)
	rootCmd.AddCommand(forwardCmd)
}