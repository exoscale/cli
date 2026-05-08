package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasReadReplicaCmd = &cobra.Command{
	Use:   "read-replica",
	Short: "Manage DBaaS read replicas",
}

func init() {
	dbaasCmd.AddCommand(dbaasReadReplicaCmd)
}
