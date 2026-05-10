package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect [name]",
	Aliases: []string{"c"},
	Short:   "Connect to a saved SSH session",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runConnect(args[0])
	},
}

func runConnect(name string) error {
	return fmt.Errorf("connect not yet implemented")
}

func init() {
	rootCmd.AddCommand(connectCmd)
}