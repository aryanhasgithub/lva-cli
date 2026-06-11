package main

import (
	"os"

	"github.com/aryanhasgithub/lva-cli/cmd"
)

func main() {
	cmd.Execute()
	if cmd.ExitWithError {
		os.Exit(1)
	}
}
