package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/aryanhasgithub/lva-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	portalPort         = 8000
	firstBootPort      = 8080
	portalPollInterval = 2 * time.Second
	portalTimeout      = 30 * time.Minute
)

// bannerCmd is hidden from `lva help` — only invoked by cli.sh.
var bannerCmd = &cobra.Command{
	Use:    "banner",
	Hidden: true,
	Short:  "Start the interactive LVA OS login shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner()
		waitForPortal()
		return runREPL()
	},
}

func init() {
	rootCmd.AddCommand(bannerCmd)
}

func printBanner() {
	fmt.Print(`
 _ __     ___       ___  ____  
| |\ \   / / \     / _ \/ ___| 
| | \ \ / / _ \   | | | \___ \ 
| |__\ V / ___ \  | |_| |___) |
|_____\_/_/   \_\  \___/|____/ 

`)
	fmt.Println("Welcome to the LVA OS command line.")
	fmt.Println("Type 'help' for available commands, 'exit' to close.")
}

// waitForPortal polls the portal health endpoint until it responds or times out.
// If the first-boot progress server is up on port 8080 instead, it prints that
// and drops into the REPL immediately without continuing to wait.
// SIGINT skips the wait entirely.
func waitForPortal() {
	skipCh := make(chan struct{}, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		fmt.Println("\nSkipping portal wait...")
		skipCh <- struct{}{}
	}()

	portalURL := fmt.Sprintf("http://127.0.0.1:%d", portalPort)
	deadline := time.Now().Add(portalTimeout)
	first := true

	if isPortalReady(portalURL) {
		printPortalReady(true)
		return
	}

	// Check for first-boot pull progress server before entering the poll loop.
	if isPortalReady(fmt.Sprintf("http://127.0.0.1:%d", firstBootPort)) {
		printFirstBootReady()
		return
	}

	ticker := time.NewTicker(portalPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-skipCh:
			fmt.Println()
			return
		case <-ticker.C:
			if time.Now().After(deadline) {
				fmt.Print("\r\033[K")
				fmt.Println("\nPortal did not become ready after 30 minutes.")
				fmt.Println()
				return
			}
			if isPortalReady(portalURL) {
				printPortalReady(first)
				return
			}
			// Mid-wait: check whether first-boot server came up.
			if isPortalReady(fmt.Sprintf("http://127.0.0.1:%d", firstBootPort)) {
				fmt.Print("\r\033[K")
				printFirstBootReady()
				return
			}
			if first {
				fmt.Print("\nPortal is not ready — please wait.")
				first = false
			} else {
				fmt.Print(".")
			}
		}
	}
}

func printPortalReady(wasWaiting bool) {
	if !wasWaiting {
		fmt.Print("\r\033[K")
	}
	ip := getHostIP()
	if ip != "" {
		fmt.Printf("\nPortal:  http://%s:%d\n\n", ip, portalPort)
	} else {
		fmt.Printf("\nPortal:  http://localhost:%d\n\n", portalPort)
	}
}

func printFirstBootReady() {
	ip := getHostIP()
	addr := "localhost"
	if ip != "" {
		addr = ip
	}
	fmt.Printf("\nFirst boot in progress — supervisor is pulling containers.\n")
	fmt.Printf("Pull progress:  http://%s:%d\n\n", addr, firstBootPort)
}

func isPortalReady(url string) bool {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

func getHostIP() string {
	sock := viper.GetString("socket")
	if sock == "" {
		sock = client.DefaultSocket
	}
	c := client.New(sock, false)
	data, err := c.NetworkInfo()
	if err != nil {
		return ""
	}

	var devices []struct {
		Interface string `json:"interface"`
		State     int    `json:"state"`
		IP4       *struct {
			Addresses []struct {
				Address string `json:"address"`
			} `json:"addresses"`
		} `json:"ip4"`
	}

	if err := json.Unmarshal(data, &devices); err != nil {
		return ""
	}

	for _, dev := range devices {
		if dev.State == 100 && dev.IP4 != nil && dev.Interface != "lo" {
			if len(dev.IP4.Addresses) > 0 {
				return dev.IP4.Addresses[0].Address
			}
		}
	}
	return ""
}

func runREPL() error {
	scanner := bufio.NewScanner(os.Stdin)

	// Catch SIGINT (Ctrl-C) — re-show prompt instead of killing the process.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for range sigCh {
			fmt.Println()
			fmt.Print("lva > ")
		}
	}()
	defer signal.Stop(sigCh)

	for {
		fmt.Print("lva > ")

		if !scanner.Scan() {
			fmt.Println()
			return nil
		}

		line := strings.TrimSpace(scanner.Text())

		switch line {
		case "":
			continue

		case "exit", "quit", "logout":
			fmt.Println("Goodbye.")
			return nil

		case "help", "?":
			rootCmd.SetArgs([]string{"help"})
			_ = rootCmd.Execute()

		default:
			input := strings.TrimPrefix(line, "lva ")
			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			rootCmd.SetArgs(parts)
			if err := rootCmd.Execute(); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				ExitWithError = false
			}
		}
	}
}
