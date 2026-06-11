package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network information and configuration",
}

// network info
var networkInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show all network interfaces with full IP info",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.NetworkInfo()
		printOrErr(data, err)
		return nil
	},
}

// network interfaces
var networkInterfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "List network interface names and states",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.NetworkInterfaces()
		printOrErr(data, err)
		return nil
	},
}

// network hostname <name>
var networkHostnameCmd = &cobra.Command{
	Use:   "hostname <new-hostname>",
	Short: "Set the system hostname",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.NetworkSetHostname(args[0])
		printOrErr(data, err)
		return nil
	},
}

// network dhcp <interface>
var networkDHCPCmd = &cobra.Command{
	Use:   "dhcp <interface>",
	Short: "Switch an interface to DHCP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.NetworkSetDHCP(args[0])
		printOrErr(data, err)
		return nil
	},
}

// network static --interface eth0 --address 192.168.1.10 --prefix 24 --gateway 192.168.1.1 --dns 1.1.1.1,8.8.8.8
var (
	staticIface   string
	staticAddress string
	staticPrefix  int
	staticGateway string
	staticDNS     string
)

var networkStaticCmd = &cobra.Command{
	Use:   "static",
	Short: "Set a static IP address on an interface",
	Long: `Set a static IP configuration on a network interface.

Example:
  lva network static --interface eth0 --address 192.168.1.10 --prefix 24 \
      --gateway 192.168.1.1 --dns 1.1.1.1,8.8.8.8`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if staticIface == "" || staticAddress == "" || staticGateway == "" || staticDNS == "" {
			fmt.Fprintln(os.Stderr, "Error: --interface, --address, --gateway, and --dns are required")
			ExitWithError = true
			return nil
		}
		if staticPrefix < 1 || staticPrefix > 32 {
			fmt.Fprintln(os.Stderr, "Error: --prefix must be between 1 and 32")
			ExitWithError = true
			return nil
		}
		dns := strings.Split(staticDNS, ",")
		for i := range dns {
			dns[i] = strings.TrimSpace(dns[i])
		}
		data, err := cli.NetworkSetStatic(staticIface, staticAddress, staticPrefix, staticGateway, dns)
		printOrErr(data, err)
		return nil
	},
}

func init() {
	networkStaticCmd.Flags().StringVar(&staticIface, "interface", "", "network interface name (e.g. eth0)")
	networkStaticCmd.Flags().StringVar(&staticAddress, "address", "", "static IPv4 address (e.g. 192.168.1.10)")
	networkStaticCmd.Flags().IntVar(&staticPrefix, "prefix", 24, "network prefix length (e.g. 24)")
	networkStaticCmd.Flags().StringVar(&staticGateway, "gateway", "", "default gateway address")
	networkStaticCmd.Flags().StringVar(&staticDNS, "dns", "", "comma-separated DNS servers (e.g. 1.1.1.1,8.8.8.8)")

	networkCmd.AddCommand(
		networkInfoCmd,
		networkInterfacesCmd,
		networkHostnameCmd,
		networkDHCPCmd,
		networkStaticCmd,
	)
	rootCmd.AddCommand(networkCmd)

}
