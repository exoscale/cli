package main

import (
	"log"

	"github.com/exoscale/cli/cmd"
)

func main() {
	if err := cmd.RootCmd.GenBashCompletionFile("bash_completion"); err != nil {
		log.Fatal(err)
	}
}
