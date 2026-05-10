package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/config"
	"github.com/sshgo/sshgo/internal/history"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View connection history",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory()
	},
}

func runHistory() error {
	histPath, err := config.DefaultHistoryPath()
	if err != nil {
		return err
	}
	h := history.NewTracker(histPath)
	entries := h.List()
	if len(entries) == 0 {
		fmt.Println("No connection history.")
		return nil
	}
	fmt.Println("Connection history:")
	for _, e := range entries {
		fmt.Printf("  %s  (connected %d times, last: %s)\n",
			e.Name, e.ConnectCount, e.LastConnected.Format("2006-01-02 15:04"))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(historyCmd)
}