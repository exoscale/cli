package cmd

import (
	"github.com/spf13/cobra"
)

var labAICmd = &cobra.Command{
	Use:   "ai",
	Short: "AI services management",
}

func init() {
	labCmd.AddCommand(labAICmd)
}
