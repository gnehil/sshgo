package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var (
	jumpHosts         []string
	jumpIdentityFiles []string
)

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
	if cfg.FindProfile(name) == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	jumps, err := buildJumpHosts(jumpHosts, jumpIdentityFiles)
	if err != nil {
		return err
	}
	if len(jumps) == 0 {
		return fmt.Errorf("use --jump to specify jump host")
	}
	applyJumpHosts(cfg, name, jumps)
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

func applyJumpHosts(cfg *config.Config, name string, jumps []config.JumpHost) {
	p := cfg.FindProfile(name)
	if p == nil {
		return
	}
	p.JumpHosts = jumps
}

func buildJumpHosts(addrs, identityFiles []string) ([]config.JumpHost, error) {
	if len(identityFiles) > len(addrs) {
		return nil, fmt.Errorf("--identity-file count (%d) exceeds --jump count (%d)", len(identityFiles), len(addrs))
	}
	jumps := make([]config.JumpHost, 0, len(addrs))
	for i, addr := range addrs {
		jh, err := parseJumpHostArg(addr)
		if err != nil {
			return nil, fmt.Errorf("jump[%d]: %w", i, err)
		}
		if i < len(identityFiles) {
			path := identityFiles[i]
			if path != "" {
				if err := config.ValidateIdentityFile(path); err != nil {
					return nil, fmt.Errorf("jump[%d]: %w", i, err)
				}
				jh.IdentityFile = path
			}
		}
		jumps = append(jumps, jh)
	}
	return jumps, nil
}

func parseJumpHostArg(arg string) (config.JumpHost, error) {
	jh := config.JumpHost{Port: 22}
	parts := strings.Split(arg, "@")
	if len(parts) == 2 {
		jh.User = parts[0]
		arg = parts[1]
	}
	colonIdx := strings.LastIndex(arg, ":")
	if colonIdx > 0 {
		port, err := strconv.Atoi(arg[colonIdx+1:])
		if err != nil {
			return jh, fmt.Errorf("invalid port %q in %q: must be a number", arg[colonIdx+1:], arg)
		}
		if port < 1 || port > 65535 {
			return jh, fmt.Errorf("invalid port %d in %q: must be 1-65535", port, arg)
		}
		jh.Port = port
		jh.Host = arg[:colonIdx]
	} else {
		jh.Host = arg
	}
	if jh.User == "" {
		jh.User = "root"
	}
	jh.Name = jh.User + "@" + jh.Host
	return jh, nil
}

func init() {
	addJumpCmd.Flags().StringArrayVar(&jumpHosts, "jump", nil, "Jump host (can be repeated for chain)")
	addJumpCmd.Flags().StringArrayVarP(&jumpIdentityFiles, "identity-file", "i", nil, "Identity file for the Nth --jump (position-matched, repeatable)")
	rootCmd.AddCommand(addJumpCmd)
}
