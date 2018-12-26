package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Diff = &cobra.Command{
	Use:   "diff",
	Short: "Get a diff summary of all services in a monorepo",
	Long:  "Get a diff summary of all services in a monorepo. It will list all services and if they have been modified, removed and added.",
	Args:  cobra.ExactValidArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running args: %s\n", args)
	},
}
