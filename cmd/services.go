package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Services = &cobra.Command{
	Use:   "service",
	Short: "Get a service summary of all services in a monorepo",
	Long:  "Get a service summary of all services in a monorepo. ",
	Args:  cobra.ExactValidArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running services args: %s\n", args)
	},
}
