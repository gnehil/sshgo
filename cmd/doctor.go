package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Audit profiles for problems (missing or insecure identity files)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor()
	},
}

func runDoctor() error {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return err
	}
	return runDoctorWithConfig(cfgPath)
}

func runDoctorWithConfig(cfgPath string) error {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}
	issues := config.Doctor(cfg)
	if len(issues) == 0 {
		fmt.Println("[OK] All profiles healthy.")
		return nil
	}
	fmt.Printf("Found %d issue(s):\n", len(issues))
	for _, i := range issues {
		loc := i.Profile
		if i.Jump != "" {
			loc = i.Profile + " " + i.Jump
		}
		fmt.Printf("  - %s: %v\n", loc, i.Err)
	}
	return fmt.Errorf("%d issue(s) found", len(issues))
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
