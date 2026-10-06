package acl

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	v3 "github.com/exoscale/egoscale/v3"
)

// Cmd is the root command for DBaaS ACL subcommands.
var Cmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage DBaaS ACL configuration",
}

// dbaasServiceType returns the type of the Database Service named name.
func dbaasServiceType(ctx context.Context, client *v3.Client, name, zone string) (string, error) {
	dbs, err := client.ListDBAASServices(ctx)
	if err != nil {
		return "", err
	}

	for _, db := range dbs.DBAASServices {
		if string(db.Name) == name {
			return string(db.Type), nil
		}
	}

	return "", fmt.Errorf("%q Database Service not found in zone %q", name, zone)
}
