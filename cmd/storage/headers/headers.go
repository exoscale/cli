package headers

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "headers",
	Short: "Manage objects HTTP headers",
}
