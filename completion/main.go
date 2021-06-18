package main

import (
	"log"
	"os"

	"github.com/exoscale/cli/cmd"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s (bash|powershell|zsh)", os.Args[0])
	}

	var err error
	switch os.Args[1] {
	case "bash":
		err = cmd.RootCmd.GenBashCompletionFile("bash_completion")

	case "powershell":
		err = cmd.RootCmd.GenPowerShellCompletionFile("powershell_completion")

	case "zsh":
		err = cmd.RootCmd.GenZshCompletionFile("zsh_completion")

	default:
		log.Fatalf("unsupported shell %q", os.Args[1])
	}

	if err != nil {
		log.Fatal(err)
	}
}
