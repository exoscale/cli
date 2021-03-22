package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var dnsCreateCmd = &cobra.Command{
	Use:     "create DOMAIN",
	Short:   "Create a domain",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		resp, err := createDomain(args[0])
		if err != nil {
			return err
		}

		if !gQuiet {
			fmt.Printf("Domain %q was created successfully\n", resp.Name)
		}

		return nil
	},
}

func createDomain(domainName string) (*egoscale.DNSDomain, error) {
	return csDNS.CreateDomain(gContext, domainName)
}

func init() {
	dnsCmd.AddCommand(dnsCreateCmd)
}
