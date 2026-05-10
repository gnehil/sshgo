package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/executor"
	"github.com/sshgo/sshgo/internal/history"
)

var recentFlag bool

var connectCmd = &cobra.Command{
	Use:     "connect [name]",
	Aliases: []string{"c"},
	Short:   "Connect to a saved SSH session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if recentFlag {
			return runConnectRecent()
		}
		if len(args) == 0 {
			return cmd.Help()
		}
		return runConnect(args[0])
	},
}

func runConnect(name string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	p := cfg.FindProfile(name)
	if p == nil {
		return fmt.Errorf("配置 %q 不存在", name)
	}
	histPath, _ := config.DefaultHistoryPath()
	h := history.NewTracker(histPath)
	h.Record(name)
	_ = h.Save()
	return executor.ExecSSH(*p)
}

func runConnectRecent() error {
	histPath, err := config.DefaultHistoryPath()
	if err != nil {
		return err
	}
	h := history.NewTracker(histPath)
	recent := h.Recent(10)
	if len(recent) == 0 {
		return fmt.Errorf("没有连接历史")
	}
	fmt.Println("Recent connections:")
	for i, r := range recent {
		fmt.Printf("  %d) %s (last: %s, count: %d)\n", i+1, r.Name, r.LastConnected.Format("2006-01-02 15:04"), r.ConnectCount)
	}
	fmt.Print("\nSelect a number: ")
	var n int
	fmt.Scanln(&n)
	if n < 1 || n > len(recent) {
		return fmt.Errorf("无效选择")
	}
	return runConnect(recent[n-1].Name)
}

func init() {
	connectCmd.Flags().BoolVar(&recentFlag, "recent", false, "Show recent connections")
	rootCmd.AddCommand(connectCmd)
}