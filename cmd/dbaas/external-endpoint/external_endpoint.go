package externalendpoint

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "external-endpoint",
	Short: "Manage DBaaS external endpoints",
}
