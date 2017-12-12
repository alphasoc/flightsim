package cmd

import (
	"fmt"

	"github.com/alphasoc/flightsim/version"
	"github.com/spf13/cobra"
)

// NewRootCommand represents the base command when called without any subcommands
func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "flightsim run|scan|version",
		Long: fmt.Sprintf("\nAlphaSOC Network Flight Simulator™ %s (https://github.com/alphasoc/flightsim)\n\n"+
			"flightsim is an application which generates malicious network traffic for security\n"+
			"teams to evaluate security controls (e.g. firewalls) and ensure that monitoring tools\n"+
			"are able to detect malicious traffic.", version.Version),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newVersionCommand())
	return cmd
}
