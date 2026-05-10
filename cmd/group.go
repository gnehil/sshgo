package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/output"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage connection groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGroupList()
	},
}

func runGroupList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Groups) == 0 {
		fmt.Println("No groups defined.")
		return nil
	}
	var rows [][]string
	for _, g := range cfg.Groups {
		count := 0
		for _, p := range cfg.Profiles {
			if p.Group == g.Name {
				count++
			}
		}
		rows = append(rows, []string{g.Name, g.Description, fmt.Sprintf("%d", count)})
	}
	return output.PrintTable(rows, []string{"NAME", "DESCRIPTION", "PROFILES"})
}

var groupAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGroupAdd(args[0])
	},
}

var groupDesc string

func runGroupAdd(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if cfg.FindGroup(name) != nil {
		return fmt.Errorf("group %q already exists", name)
	}
	cfg.Groups = append(cfg.Groups, config.Group{Name: name, Description: groupDesc})
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGroupDelete(args[0])
	},
}

func runGroupDelete(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	found := false
	for i, g := range cfg.Groups {
		if g.Name == name {
			cfg.Groups = append(cfg.Groups[:i], cfg.Groups[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("group %q not found", name)
	}
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

func init() {
	groupAddCmd.Flags().StringVar(&groupDesc, "description", "", "Group description")
	groupCmd.AddCommand(groupListCmd, groupAddCmd, groupDeleteCmd)
	rootCmd.AddCommand(groupCmd)
}