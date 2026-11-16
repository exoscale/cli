package readreplica

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "read-replica",
	Short: "Manage Database Service read replicas",
}
