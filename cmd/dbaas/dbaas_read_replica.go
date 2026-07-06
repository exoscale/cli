package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasReadReplicaCmd = &cobra.Command{
	Use:   "read-replica",
	Short: "Manage Database Service read replicas",
}

func init() {
	dbaasCmd.AddCommand(dbaasReadReplicaCmd)
}
