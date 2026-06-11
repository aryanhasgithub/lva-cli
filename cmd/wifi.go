package cmd

import (
	"github.com/spf13/cobra"
)

// wifiCmd is a subcommand of networkCmd, not rootCmd.
var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "WiFi network scanning and connection management",
}

// network wifi scan --interface wlan0
var wifiScanIface string

var wifiScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for available WiFi networks",
	Long: `Trigger a WiFi scan and list visible access points sorted by signal strength.

Example:
  lva network wifi scan --interface wlan0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiScanIface == "" {
			wifiScanIface = "wlan0"
		}
		data, err := cli.WifiScan(wifiScanIface)
		printOrErr(data, err)
		return nil
	},
}

// network wifi connect --interface wlan0 --ssid MyNetwork [--password secret]
var (
	wifiConnectIface    string
	wifiConnectSSID     string
	wifiConnectPassword string
)

var wifiConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a WiFi network",
	Long: `Connect to a WiFi network by SSID.
Omit --password for open networks.

Examples:
  lva network wifi connect --interface wlan0 --ssid MyNetwork --password secret
  lva network wifi connect --interface wlan0 --ssid OpenNetwork`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiConnectIface == "" {
			wifiConnectIface = "wlan0"
		}
		if wifiConnectSSID == "" {
			return runError("--ssid is required")
		}

		var password *string
		if cmd.Flags().Changed("password") {
			password = &wifiConnectPassword
		}

		data, err := cli.WifiConnect(wifiConnectIface, wifiConnectSSID, password)
		printOrErr(data, err)
		return nil
	},
}

// network wifi disconnect --interface wlan0
var wifiDisconnectIface string

var wifiDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from the active WiFi connection",
	Long: `Disconnect the active connection on a WiFi interface.

Example:
  lva network wifi disconnect --interface wlan0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wifiDisconnectIface == "" {
			wifiDisconnectIface = "wlan0"
		}
		data, err := cli.WifiDisconnect(wifiDisconnectIface)
		printOrErr(data, err)
		return nil
	},
}

func init() {
	wifiScanCmd.Flags().StringVar(&wifiScanIface, "interface", "wlan0",
		"WiFi interface to scan on")

	wifiConnectCmd.Flags().StringVar(&wifiConnectIface, "interface", "wlan0",
		"WiFi interface to connect on")
	wifiConnectCmd.Flags().StringVar(&wifiConnectSSID, "ssid", "",
		"SSID of the network to connect to")
	wifiConnectCmd.Flags().StringVar(&wifiConnectPassword, "password", "",
		"WiFi password (omit for open networks)")

	wifiDisconnectCmd.Flags().StringVar(&wifiDisconnectIface, "interface", "wlan0",
		"WiFi interface to disconnect")

	wifiCmd.AddCommand(wifiScanCmd, wifiConnectCmd, wifiDisconnectCmd)

	// wifi lives under network, not root
	networkCmd.AddCommand(wifiCmd)
}

// runError is a small helper to print an error and set the exit flag.
func runError(msg string) error {
	ExitWithError = true
	return nil
}