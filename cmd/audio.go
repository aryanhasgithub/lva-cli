package cmd

import (
	"github.com/spf13/cobra"
)

var audioCmd = &cobra.Command{
	Use:   "audio",
	Short: "Audio device information",
}

// audio devices
var audioDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available input and output audio devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.AudioDevices()
		printOrErr(data, err)
		return nil
	},
}

func init() {
	audioCmd.AddCommand(audioDevicesCmd)
	rootCmd.AddCommand(audioCmd)
}
