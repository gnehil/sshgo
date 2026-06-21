package cmd

import (
	"fmt"
	"os"
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
	jumps, err := buildJumpHosts(jumpHosts, jumpIdentityFiles)
	if err != nil {
		return err
	}
	if len(jumps) == 0 {
		return fmt.Errorf("use --jump to specify jump host")
	}
	if err := applyJumpHosts(cfg, name, jumps); err != nil {
		return err
	}
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

func applyJumpHosts(cfg *config.Config, name string, jumps []config.JumpHost) error {
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("profile %q not found", name)
	}
	p.JumpHosts = jumps
	return nil
}

func buildJumpHosts(addrs, identityFiles []string) ([]config.JumpHost, error) {
	if len(identityFiles) > len(addrs) {
		return nil, fmt.Errorf("--identity-file count (%d) exceeds --jump count (%d)", len(identityFiles), len(addrs))
	}
	jumps := make([]config.JumpHost, 0, len(addrs))
	for i, addr := range addrs {
		jh := parseJumpHostArg(addr)
		if i < len(identityFiles) {
			path := identityFiles[i]
			if path != "" {
				if _, err := os.Stat(path); err != nil {
					return nil, fmt.Errorf("jump[%d] identity file: %w", i, err)
				}
				jh.IdentityFile = path
			}
		}
		jumps = append(jumps, jh)
	}
	return jumps, nil
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
	addJumpCmd.Flags().StringArrayVarP(&jumpIdentityFiles, "identity-file", "i", nil, "Identity file for the Nth --jump (position-matched, repeatable)")
	rootCmd.AddCommand(addJumpCmd)
}
