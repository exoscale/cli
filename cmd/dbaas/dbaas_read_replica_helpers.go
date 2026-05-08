package dbaas

import (
	"context"
	"fmt"
	"sort"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasServiceWithZone struct {
	Service v3.DBAASServiceCommon
	Zone    string
}

func dbaasListServicesAllZones(ctx context.Context) ([]dbaasServiceWithZone, error) {
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, "")
	if err != nil {
		return nil, err
	}

	out := make([]dbaasServiceWithZone, 0)

	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		zonalClient := client.WithEndpoint(zone.APIEndpoint)

		list, err := zonalClient.ListDBAASServices(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Database Services in zone %s: %w", zone, err)
		}

		for _, svc := range list.DBAASServices {
			out = append(out, dbaasServiceWithZone{
				Service: svc,
				Zone:    string(zone.Name),
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Zone == out[j].Zone {
			return string(out[i].Service.Name) < string(out[j].Service.Name)
		}
		return out[i].Zone < out[j].Zone
	})

	return out, nil
}

func dbaasFindServiceByNameAllZones(ctx context.Context, name string) (dbaasServiceWithZone, error) {
	services, err := dbaasListServicesAllZones(ctx)
	if err != nil {
		return dbaasServiceWithZone{}, err
	}

	var matches []dbaasServiceWithZone

	for _, svc := range services {
		if string(svc.Service.Name) == name {
			matches = append(matches, svc)
		}
	}

	switch len(matches) {
	case 0:
		return dbaasServiceWithZone{}, fmt.Errorf("%q Database Service not found", name)
	case 1:
		return matches[0], nil
	default:
		return dbaasServiceWithZone{}, fmt.Errorf("%q multiple Database Services found", name)
	}
}

func dbaasGetReadReplicaIntegrationForReplica(service v3.DBAASServiceCommon) *v3.DBAASIntegration {
	serviceName := string(service.Name)

	for _, integration := range service.Integrations {
		if integration.Type == "read_replica" && integration.Dest == serviceName {
			replicaIntegration := integration
			return &replicaIntegration
		}
	}

	return nil
}

func dbaasActiveReadReplicaNamesForPrimary(service v3.DBAASServiceCommon) []string {
	serviceName := string(service.Name)
	replicaNames := make([]string, 0)

	for _, integration := range service.Integrations {
		if integration.Type != "read_replica" {
			continue
		}
		if !utils.DefaultBool(integration.ISActive, false) {
			continue
		}
		if integration.Source != serviceName || integration.Dest == "" {
			continue
		}

		replicaNames = append(replicaNames, integration.Dest)
	}

	sort.Strings(replicaNames)

	return replicaNames
}

func dbaasReadReplicaSupportedServiceType(serviceType string) bool {
	switch serviceType {
	case "pg", "mysql":
		return true
	default:
		return false
	}
}

func dbaasReadReplicaClientForZone(ctx context.Context, zone string) (*v3.Client, error) {
	return exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
}
