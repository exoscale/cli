// Package internal holds helpers shared by the dbaas command packages.
package internal

import (
	"context"
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

// GetV3 returns the Database Service named name in zone.
func GetV3(ctx context.Context, name, zone string) (v3.DBAASServiceCommon, error) {
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
	if err != nil {
		return v3.DBAASServiceCommon{}, err
	}

	dbs, err := client.ListDBAASServices(ctx)
	if err != nil {
		return v3.DBAASServiceCommon{}, err
	}

	for _, db := range dbs.DBAASServices {
		if string(db.Name) == name {
			return db, nil
		}
	}

	return v3.DBAASServiceCommon{}, fmt.Errorf("%q Database Service not found in zone %q", name, zone)
}
