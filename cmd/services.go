package cmd

import (
	"github.com/spf13/cobra"
)

// Services returns the services command without flags or Run args
func Services() *cobra.Command {
	return &cobra.Command{
		Use:   "services",
		Short: "Get a summary of all services in a monorepo",
		Long:  "Get a summary of all services in a monorepo. ",
		Args:  cobra.ExactValidArgs(0),
	}
}
