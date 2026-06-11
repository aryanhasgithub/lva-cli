package cmd

import (
	"github.com/spf13/cobra"
)

var updatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Check and apply component and OS updates",
}

// updates check
var updatesCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Compare local image versions against the remote version manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.UpdatesCheck()
		printOrErr(data, err)
		return nil
	},
}

// updates versions
var updatesVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Show current local OCI version labels for all components",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.UpdatesVersions()
		printOrErr(data, err)
		return nil
	},
}

// updates os
var updatesOSCmd = &cobra.Command{
	Use:   "os",
	Short: "Check the version manifest for a new LVA OS bundle",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.UpdatesOSInfo()
		printOrErr(data, err)
		return nil
	},
}

// updates apply <component>
var updatesApplyCmd = &cobra.Command{
	Use:   "apply <component>",
	Short: "Pull the latest image for a specific component",
	Long: `Pull and hot-swap the image for a single managed component.

Valid component names: lva, lva-audio, lva-portal`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.UpdatesUpdateComponent(args[0])
		printOrErr(data, err)
		return nil
	},
}

func init() {
	updatesApplyCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return managedContainers(), cobra.ShellCompDirectiveNoFileComp
	}

	updatesCmd.AddCommand(
		updatesCheckCmd,
		updatesVersionsCmd,
		updatesOSCmd,
		updatesApplyCmd,
	)
	rootCmd.AddCommand(updatesCmd)
}
