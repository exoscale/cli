package database

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "database",
	Short: "Manage DBaaS databases",
}
