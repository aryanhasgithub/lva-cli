// Package cmd contains all CLI subcommands for lva.
package cmd

import (
	"fmt"
	"os"

	"github.com/aryanhasgithub/lva-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ExitWithError is set to true when any command fails, so main() can exit(1).
var ExitWithError bool

var (
	cfgFile    string
	socketPath string
	rawJSON    bool
	logLevel   string
)

// cli is the shared client instance, built in PersistentPreRun.
var cli *client.Client

var rootCmd = &cobra.Command{
	Use:   "lva",
	Short: "LVA OS command-line interface",
	Long:  `Command-line interface for interacting with the LVA OS supervisor.`,
	// Require a subcommand — don't run anything at the root level.
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("subcommand is required — run 'lva help'")
	},
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ExitWithError = true
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"optional config file (default: $HOME/.lva.yaml)")
	rootCmd.PersistentFlags().StringVar(&socketPath, "socket", client.DefaultSocket,
		"path to supervisor Unix socket")
	rootCmd.PersistentFlags().BoolVar(&rawJSON, "raw-json", false,
		"output raw JSON from the API")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn",
		"log level (debug, info, warn, error)")

	// Allow all flags from LVA_ env vars (e.g. LVA_SOCKET, LVA_RAW_JSON)
	viper.SetEnvPrefix("LVA")
	viper.AutomaticEnv()

	_ = viper.BindPFlag("socket", rootCmd.PersistentFlags().Lookup("socket"))
	_ = viper.BindPFlag("raw_json", rootCmd.PersistentFlags().Lookup("raw-json"))

	// Build the shared client before any subcommand runs.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		sock := viper.GetString("socket")
		if sock == "" {
			sock = client.DefaultSocket
		}
		cli = client.New(sock, rawJSON || viper.GetBool("raw_json"))
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.SetConfigName(".lva")
	}
	_ = viper.ReadInConfig()
}

// printOrErr is a helper used by every command — print result or set error flag.
func printOrErr(data []byte, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		ExitWithError = true
		return
	}
	if err := cli.PrintJSON(data); err != nil {
		fmt.Fprintln(os.Stderr, "Error formatting output:", err)
		ExitWithError = true
	}
}
