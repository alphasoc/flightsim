package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version string to be filled during a build process
var Version = ""

// NewRootCommand represents the base command when called without any subcommands
func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "flightsim",
		Long: fmt.Sprintf("AlphaSOC Network Flight Simulator™ (https://github.com/alphasoc/flightsim)\n\n" +
			"flightsim is an application which generates malicious network traffic for security\n" +
			"teams to evaluate security controls (e.g. firewalls) and ensure that monitoring tools\n" +
			"are able to detect malicious traffic."),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(newRunCommand())
	// Set version (if non-empty then --version will be available)
	cmd.Version = Version

	return cmd
}
