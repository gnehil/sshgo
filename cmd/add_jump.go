package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var jumpHosts []string

var addJumpCmd = &cobra.Command{
	Use:   "add-jump <name>",
	Short: "Add jump host(s) to a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddJump(args[0])
	},
}

func runAddJump(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	var jumps []config.JumpHost
	for _, j := range jumpHosts {
		jumps = append(jumps, parseJumpHostArg(j))
	}
	if len(jumps) == 0 {
		return fmt.Errorf("use --jump to specify jump host")
	}
	p.JumpHosts = jumps
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

func parseJumpHostArg(arg string) config.JumpHost {
	jh := config.JumpHost{Port: 22}
	parts := strings.Split(arg, "@")
	if len(parts) == 2 {
		jh.User = parts[0]
		arg = parts[1]
	}
	colonIdx := strings.LastIndex(arg, ":")
	if colonIdx > 0 {
		p, _ := strconv.Atoi(arg[colonIdx+1:])
		if p > 0 {
			jh.Port = p
		}
		jh.Host = arg[:colonIdx]
	} else {
		jh.Host = arg
	}
	if jh.User == "" {
		jh.User = "root"
	}
	jh.Name = jh.User + "@" + jh.Host
	return jh
}

func init() {
	addJumpCmd.Flags().StringArrayVar(&jumpHosts, "jump", nil, "Jump host (can be repeated for chain)")
	rootCmd.AddCommand(addJumpCmd)
}