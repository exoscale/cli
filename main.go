package main

import (
	"log"

	"github.com/exoscale/cli/cmd"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
