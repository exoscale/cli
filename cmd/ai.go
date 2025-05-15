package cmd

import (
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:        "ai",
	Short:      "AI services management",
	Aliases:    []string{"a"},
	SuggestFor: []string{"ai", "vm"},
}

func init() {
	RootCmd.AddCommand(aiCmd)
}
