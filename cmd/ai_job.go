package cmd

import (
	"github.com/spf13/cobra"
)

var gpuInstanceTypeFamilies = []string{
	"gpu",
	"gpu2",
	"gpu3",
}

var gpuInstanceTypeSizes = []string{
	"medium",
	"large",
	"huge",
}

var aiJobCmd = &cobra.Command{
	Use:     "job",
	Short:   "AI jobs management",
	Aliases: []string{"j"},
}

func init() {
	aiCmd.AddCommand(aiJobCmd)
}
