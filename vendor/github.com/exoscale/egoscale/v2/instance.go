package v2

import (
	"context"
	"net"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// Instance represents a Compute instance.
type Instance struct {
	AntiAffinityGroupIDs []string
	CreatedAt            time.Time
	DiskSize             int64
	ElasticIPIDs         []string
	ID                   string
	IPv6Address          net.IP
	IPv6Enabled          bool
	InstanceTypeID       string
	ManagerID            string
	Name                 string
	PrivateNetworkIDs    []string
	PublicIPAddress      net.IP
	SSHKey               string
	SecurityGroupIDs     []string
	SnapshotIDs          []string
	State                string
	TemplateID           string
	UserData             string

	c    *Client
	zone string
}

func instanceFromAPI(client *Client, zone string, i *papi.Instance) *Instance {
	return &Instance{
		AntiAffinityGroupIDs: func() []string {
			ids := make([]string, 0)

			if i.AntiAffinityGroups != nil {
				for _, item := range *i.AntiAffinityGroups {
					item := item
					ids = append(ids, *item.Id)
				}
			}

			return ids
		}(),
		CreatedAt: *i.CreatedAt,
		DiskSize:  *i.DiskSize,
		ElasticIPIDs: func() []string {
			ids := make([]string, 0)

			if i.ElasticIps != nil {
				for _, item := range *i.ElasticIps {
					item := item
					ids = append(ids, *item.Id)
				}
			}

			return ids
		}(),
		ID: *i.Id,
		IPv6Address: func() net.IP {
			if i.Ipv6Address != nil {
				return net.ParseIP(*i.Ipv6Address)
			}
			return nil
		}(),
		IPv6Enabled: func() bool {
			return i.Ipv6Address != nil
		}(),
		InstanceTypeID: *i.InstanceType.Id,
		ManagerID: func() string {
			if i.Manager != nil {
				return papi.OptionalString(i.Manager.Id)
			}
			return ""
		}(),
		Name: *i.Name,
		PrivateNetworkIDs: func() []string {
			ids := make([]string, 0)

			if i.PrivateNetworks != nil {
				for _, item := range *i.PrivateNetworks {
					item := item
					ids = append(ids, *item.Id)
				}
			}

			return ids
		}(),
		PublicIPAddress: net.ParseIP(papi.OptionalString(i.PublicIp)),
		SSHKey: func() string {
			key := ""
			if i.SshKey != nil {
				key = papi.OptionalString(i.SshKey.Name)
			}
			return key
		}(),
		SecurityGroupIDs: func() []string {
			ids := make([]string, 0)

			if i.SecurityGroups != nil {
				for _, item := range *i.SecurityGroups {
					item := item
					ids = append(ids, *item.Id)
				}
			}

			return ids
		}(),
		SnapshotIDs: func() []string {
			ids := make([]string, 0)

			if i.Snapshots != nil {
				for _, item := range *i.Snapshots {
					item := item
					ids = append(ids, *item.Id)
				}
			}

			return ids
		}(),
		State:      *i.State,
		TemplateID: *i.Template.Id,
		UserData:   papi.OptionalString(i.UserData),

		c:    client,
		zone: zone,
	}
}

func (i Instance) get(ctx context.Context, client *Client, zone, id string) (interface{}, error) {
	return client.GetInstance(ctx, zone, id)
}

// AntiAffinityGroups returns the list of Anti-Affinity Groups applied to the Compute instance.
func (i *Instance) AntiAffinityGroups(ctx context.Context) ([]*AntiAffinityGroup, error) {
	res, err := i.c.fetchFromIDs(ctx, i.zone, i.AntiAffinityGroupIDs, new(AntiAffinityGroup))
	return res.([]*AntiAffinityGroup), err
}

// ElasticIPs returns the list of Elastic IPs attached to the Compute instance.
func (i *Instance) ElasticIPs(ctx context.Context) ([]*ElasticIP, error) {
	res, err := i.c.fetchFromIDs(ctx, i.zone, i.ElasticIPIDs, new(ElasticIP))
	return res.([]*ElasticIP), err
}

// PrivateNetworks returns the list of Private Networks attached to the Compute instance.
func (i *Instance) PrivateNetworks(ctx context.Context) ([]*PrivateNetwork, error) {
	res, err := i.c.fetchFromIDs(ctx, i.zone, i.PrivateNetworkIDs, new(PrivateNetwork))
	return res.([]*PrivateNetwork), err
}

// SecurityGroups returns the list of Security Groups attached to the Compute instance.
func (i *Instance) SecurityGroups(ctx context.Context) ([]*SecurityGroup, error) {
	res, err := i.c.fetchFromIDs(ctx, i.zone, i.SecurityGroupIDs, new(SecurityGroup))
	return res.([]*SecurityGroup), err
}

// CreateSnapshot creates a Snapshot of the Compute instance storage volume.
func (i *Instance) CreateSnapshot(ctx context.Context) (*Snapshot, error) {
	resp, err := i.c.CreateSnapshotWithResponse(apiv2.WithZone(ctx, i.zone), i.ID)
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(i.c.timeout).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return i.c.GetSnapshot(ctx, i.zone, *res.(*papi.Reference).Id)
}

// RevertToSnapshot reverts the Compute instance storage volume to the specified Snapshot.
func (i *Instance) RevertToSnapshot(ctx context.Context, snapshot *Snapshot) error {
	resp, err := i.c.RevertInstanceToSnapshotWithResponse(
		apiv2.WithZone(ctx, i.zone),
		i.ID,
		papi.RevertInstanceToSnapshotJSONRequestBody{Id: snapshot.ID})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateInstance creates a Compute instance in the specified zone.
func (c *Client) CreateInstance(ctx context.Context, zone string, instance *Instance) (*Instance, error) {
	resp, err := c.CreateInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateInstanceJSONRequestBody{
			AntiAffinityGroups: func() *[]papi.AntiAffinityGroup {
				if l := len(instance.AntiAffinityGroupIDs); l > 0 {
					list := make([]papi.AntiAffinityGroup, l)
					for i, v := range instance.AntiAffinityGroupIDs {
						v := v
						list[i] = papi.AntiAffinityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
			DiskSize:     instance.DiskSize,
			InstanceType: papi.InstanceType{Id: &instance.InstanceTypeID},
			Ipv6Enabled:  &instance.IPv6Enabled,
			Name:         &instance.Name,
			SecurityGroups: func() *[]papi.SecurityGroup {
				if l := len(instance.SecurityGroupIDs); l > 0 {
					list := make([]papi.SecurityGroup, l)
					for i, v := range instance.SecurityGroupIDs {
						v := v
						list[i] = papi.SecurityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
			SshKey: func() *papi.SshKey {
				if instance.SSHKey != "" {
					return &papi.SshKey{Name: &instance.SSHKey}
				}
				return nil
			}(),
			Template: papi.Template{Id: &instance.TemplateID},
			UserData: func() *string {
				if instance.UserData != "" {
					return &instance.UserData
				}
				return nil
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetInstance(ctx, zone, *res.(*papi.Reference).Id)
}

// ListInstances returns the list of existing Compute instances in the specified zone.
func (c *Client) ListInstances(ctx context.Context, zone string) ([]*Instance, error) {
	list := make([]*Instance, 0)

	resp, err := c.ListInstancesWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Instances != nil {
		for i := range *resp.JSON200.Instances {
			list = append(list, instanceFromAPI(c, zone, &(*resp.JSON200.Instances)[i]))
		}
	}

	return list, nil
}

// GetInstance returns the Instance  corresponding to the specified ID in the specified zone.
func (c *Client) GetInstance(ctx context.Context, zone, id string) (*Instance, error) {
	resp, err := c.GetInstanceWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return instanceFromAPI(c, zone, resp.JSON200), nil
}

// UpdateInstance updates the specified Compute instance in the specified zone.
func (c *Client) UpdateInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.UpdateInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		instance.ID,
		papi.UpdateInstanceJSONRequestBody{
			Name: func() *string {
				if instance.Name != "" {
					return &instance.Name
				}
				return nil
			}(),
			UserData: func() *string {
				if instance.UserData != "" {
					return &instance.UserData
				}
				return nil
			}(),
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteInstance deletes the specified Compute instance in the specified zone.
func (c *Client) DeleteInstance(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteInstanceWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
