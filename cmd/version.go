package cmd

import (
	"fmt"

	"github.com/alphasoc/flightsim/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Run:   printversion,
	}
}

func printversion(cmd *cobra.Command, args []string) {
	fmt.Printf("flightsim version %s\n", version.Version)
}
