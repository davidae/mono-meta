package cmd

import (
	"github.com/spf13/cobra"
)

// Diff returns the diff command without flags or Run args
func Diff() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Get a diff summary of all services in a monorepo between two references",
		Long:  "Get a diff summary of all services in a monorepo. It will list all services and if they have been modified, removed and added.",
		Args:  cobra.ExactValidArgs(0),
	}
}
