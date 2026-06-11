package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Control the LVA OS host system",
}

// system reboot
var systemRebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot the host system",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.SystemReboot()
		printOrErr(data, err)
		return nil
	},
}

// system poweroff
var systemPoweroffCmd = &cobra.Command{
	Use:   "poweroff",
	Short: "Power off the host system",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.SystemPoweroff()
		printOrErr(data, err)
		return nil
	},
}

// system health
var systemHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show supervisor and container health",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := cli.SystemHealth()
		printOrErr(data, err)
		return nil
	},
}

// system os-update --bundle-url <url>
var osUpdateBundleURL string

var systemOSUpdateCmd = &cobra.Command{
	Use:   "os-update",
	Short: "Trigger a RAUC OTA bundle install",
	Long: `Triggers a RAUC A/B OTA update on the host.
The supervisor blocks until RAUC completes. After success, reboot to apply.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if osUpdateBundleURL == "" {
			fmt.Fprintln(os.Stderr, "Error: --bundle-url is required")
			ExitWithError = true
			return nil
		}
		data, err := cli.SystemOSUpdate(osUpdateBundleURL)
		printOrErr(data, err)
		return nil
	},
}

// system update-stream <name>  — streams SSE from /containers/<name>/update/stream
var systemUpdateStreamCmd = &cobra.Command{
	Use:   "update-stream <container>",
	Short: "Stream live progress of a container image update",
	Long: `Pulls the latest image for a container and streams each step to stdout.
Useful for watching long pulls on slow connections.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return streamContainerUpdate(args[0])
	},
}

func init() {
	systemOSUpdateCmd.Flags().StringVar(&osUpdateBundleURL, "bundle-url", "",
		"URL of the RAUC bundle to install")

	systemUpdateStreamCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return managedContainers(), cobra.ShellCompDirectiveNoFileComp
	}

	systemCmd.AddCommand(
		systemRebootCmd,
		systemPoweroffCmd,
		systemHealthCmd,
		systemOSUpdateCmd,
		systemUpdateStreamCmd,
	)
	rootCmd.AddCommand(systemCmd)

	// Keep context import used.
}

// streamContainerUpdate connects to the supervisor SSE endpoint over the Unix
// socket and prints progress lines until the stream closes or an error arrives.
func streamContainerUpdate(name string) error {
	sock := viper.GetString("socket")
	if sock == "" {
		sock = "/run/lva/supervisor.sock"
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return fmt.Errorf("cannot connect to supervisor socket: %w", err)
	}
	defer conn.Close()

	path := fmt.Sprintf("/containers/%s/update/stream", name)
	rawReq := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: lva\r\nAccept: text/event-stream\r\nConnection: close\r\n\r\n",
		path,
	)
	if _, err := conn.Write([]byte(rawReq)); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	scanner := bufio.NewScanner(conn)

	// Skip HTTP response headers — stop at the blank line.
	for scanner.Scan() {
		if scanner.Text() == "" {
			break
		}
	}

	// Read SSE events line by line.
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")

		var event struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			fmt.Println(payload)
			continue
		}

		switch event.Type {
		case "log":
			fmt.Println("  →", event.Message)
		case "success":
			fmt.Println("✓", event.Message)
			return nil
		case "error":
			fmt.Fprintln(os.Stderr, "✗", event.Message)
			ExitWithError = true
			return nil
		default:
			fmt.Println(payload)
		}
	}
	return nil
}
