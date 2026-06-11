package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage LVA containers",
}

// containers list
var containersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed containers and their state",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerList()
		printOrErr(data, err)
		return nil
	},
}

// containers start <name>
var containersStartCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerStart(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers stop <name>
var containersStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerStop(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers restart <name>
var containersRestartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "Restart a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerRestart(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers state <name>
var containersStateCmd = &cobra.Command{
	Use:   "state <name>",
	Short: "Get the current state of a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerState(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers update <name>
var containersUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Pull the latest image and recreate a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerUpdate(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers stats <name>
var containersStatsCmd = &cobra.Command{
	Use:   "stats <name>",
	Short: "Show CPU and memory stats for a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.ContainerStats(args[0])
		printOrErr(data, err)
		return nil
	},
}

// containers logs <name> [--tail N]
var logsTail int

var containersLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Fetch recent log lines from a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if logsTail < 1 {
			fmt.Fprintln(os.Stderr, "Error: --tail must be >= 1")
			ExitWithError = true
			return nil
		}
		data, err := cli.ContainerLogs(args[0], logsTail)
		printOrErr(data, err)
		return nil
	},
}

func init() {
	containersLogsCmd.Flags().IntVarP(&logsTail, "tail", "n", 100,
		"number of log lines to return")

	containersCmd.AddCommand(
		containersListCmd,
		containersStartCmd,
		containersStopCmd,
		containersRestartCmd,
		containersStateCmd,
		containersUpdateCmd,
		containersStatsCmd,
		containersLogsCmd,
	)
	rootCmd.AddCommand(containersCmd)

	// Teach cobra about the valid container names for shell completion.
	for _, c := range []*cobra.Command{
		containersStartCmd,
		containersStopCmd,
		containersRestartCmd,
		containersStateCmd,
		containersUpdateCmd,
		containersStatsCmd,
		containersLogsCmd,
	} {
		_ = c.RegisterFlagCompletionFunc("", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return managedContainers(), cobra.ShellCompDirectiveNoFileComp
		})
		c.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return managedContainers(), cobra.ShellCompDirectiveNoFileComp
		}
	}

}

// managedContainers returns the static list of containers for shell completion.
// Mirrors MANAGED_CONTAINERS in lva-supervisor const.py.
func managedContainers() []string {
	return []string{"lva", "lva-audio", "lva-portal"}
}
