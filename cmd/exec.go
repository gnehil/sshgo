package cmd

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/cobra"
	"github.com/sshgo/sshgo/internal/batch"
)

var execGroup string

var execCmd = &cobra.Command{
	Use:   "exec <pattern> <command>",
	Short: "Execute command on multiple servers",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		command := strings.Join(args[1:], " ")
		return runExec(pattern, command)
	},
}

func runExec(pattern, command string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	targets := batch.FindTargets(cfg, pattern, execGroup)
	if len(targets) == 0 {
		return fmt.Errorf("没有找到匹配的服务器")
	}
	var execTargets []batch.Target
	for _, t := range targets {
		execTargets = append(execTargets, batch.Target{Name: t.Name, Profile: t, Command: command})
	}
	results := batch.ExecuteAll(execTargets)
	success, failed := 0, 0
	for _, r := range results {
		fmt.Printf("\n=== %s ===\n", r.Name)
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", r.Err)
			failed++
		} else {
			fmt.Println(r.Output)
			success++
		}
	}
	fmt.Printf("\nTotal: %d succeeded, %d failed\n", success, failed)
	return nil
}

func init() {
	execCmd.Flags().StringVar(&execGroup, "group", "", "Execute on all servers in a group")
	rootCmd.AddCommand(execCmd)
}