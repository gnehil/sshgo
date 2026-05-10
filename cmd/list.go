package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/output"
)

var (
	listGroup  string
	listSort   string
	listFormat string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List SSH connection profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func runList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	profiles := cfg.Profiles
	if listGroup != "" {
		var filtered []config.Profile
		for _, p := range profiles {
			if p.Group == listGroup {
				filtered = append(filtered, p)
			}
		}
		profiles = filtered
	}
	if listSort == "name" {
		for i := 0; i < len(profiles); i++ {
			for j := i + 1; j < len(profiles); j++ {
				if profiles[i].Name > profiles[j].Name {
					profiles[i], profiles[j] = profiles[j], profiles[i]
				}
			}
		}
	}
	switch listFormat {
	case "json":
		return output.PrintJSON(profiles)
	case "table", "":
		if len(profiles) == 0 {
			fmt.Println("No profiles found.")
			return nil
		}
		var rows [][]string
		for _, p := range profiles {
			rows = append(rows, []string{
				p.Name,
				fmt.Sprintf("%s:%d", p.Host, p.Port),
				p.User,
				p.Group,
			})
		}
		return output.PrintTable(rows, []string{"NAME", "HOST", "USER", "GROUP"})
	default:
		return fmt.Errorf("unknown format: %s", listFormat)
	}
}

func init() {
	listCmd.Flags().StringVar(&listGroup, "group", "", "Filter by group")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by: name")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "Output format: table|json")
	rootCmd.AddCommand(listCmd)
}