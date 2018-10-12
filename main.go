package main

import (
	"log"

	"github.com/exoscale/cli/cmd"
)

var (
	// version comes from the tag (during the build)
	version = "dev"
	// commit represents the HEAD commit (during the build)
	commit = "n/a"
)

func main() {
	log.SetFlags(0)
	cmd.Execute(version, commit)
}
