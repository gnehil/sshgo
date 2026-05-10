package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/sshconfig"
)

var (
	importFile     string
	importOverwrite bool
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import profiles from ~/.ssh/config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImport()
	},
}

func runImport() error {
	sourcePath := importFile
	if sourcePath == "" {
		sourcePath = config.ExpandTilde("~/.ssh/config")
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", sourcePath)
	}
	profiles, err := sshconfig.ParseSSHConfig(sourcePath)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}
	if len(profiles) == 0 {
		fmt.Println("No profiles found in SSH config.")
		return nil
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	added, skipped := 0, 0
	for _, p := range profiles {
		if cfg.FindProfile(p.Name) != nil && !importOverwrite {
			skipped++
			continue
		}
		p.Group = "Imported from SSH config"
		if cfg.FindGroup("Imported from SSH config") == nil {
			cfg.Groups = append(cfg.Groups, config.Group{Name: "Imported from SSH config", Description: "Auto-imported from OpenSSH config"})
		}
		cfg.AddProfile(p)
		added++
	}
	cfgPath, _ := config.DefaultConfigPath()
	return config.SaveConfig(cfgPath, cfg)
}

func init() {
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "Source file (default: ~/.ssh/config)")
	importCmd.Flags().BoolVarP(&importOverwrite, "overwrite", "w", false, "Overwrite existing profiles")
	rootCmd.AddCommand(importCmd)
}