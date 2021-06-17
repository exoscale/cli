package v2

import (
	"context"
	"net"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// InstanceManager represents a Compute instance manager.
type InstanceManager struct {
	ID   string
	Type string
}

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
	Labels               map[string]string `reset:"labels"`
	Manager              *InstanceManager
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
		Labels: func() map[string]string {
			if i.Labels != nil {
				return i.Labels.AdditionalProperties
			}
			return nil
		}(),
		Manager: func() *InstanceManager {
			if i.Manager != nil {
				return &InstanceManager{
					ID:   *i.Manager.Id,
					Type: string(*i.Manager.Type),
				}
			}
			return nil
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
				key = *i.SshKey.Name
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
		State:      string(*i.State),
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

// AttachElasticIP attaches the Compute instance to the specified Elastic IP.
func (i *Instance) AttachElasticIP(ctx context.Context, elasticIP *ElasticIP) error {
	resp, err := i.c.AttachInstanceToElasticIpWithResponse(
		apiv2.WithZone(ctx, i.zone), elasticIP.ID, papi.AttachInstanceToElasticIpJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
		})
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

// AttachPrivateNetwork attaches the Compute instance to the specified Private Network.
func (i *Instance) AttachPrivateNetwork(ctx context.Context, privateNetwork *PrivateNetwork, address net.IP) error {
	resp, err := i.c.AttachInstanceToPrivateNetworkWithResponse(
		apiv2.WithZone(ctx, i.zone), privateNetwork.ID, papi.AttachInstanceToPrivateNetworkJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
			Ip: func() *string {
				if len(address) > 0 {
					ip := address.String()
					return &ip
				}
				return nil
			}(),
		})
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

// AttachSecurityGroup attaches the Compute instance to the specified Security Group.
func (i *Instance) AttachSecurityGroup(ctx context.Context, securityGroup *SecurityGroup) error {
	resp, err := i.c.AttachInstanceToSecurityGroupWithResponse(
		apiv2.WithZone(ctx, i.zone), securityGroup.ID, papi.AttachInstanceToSecurityGroupJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
		})
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

// CreateSnapshot creates a Snapshot of the Compute instance storage volume.
func (i *Instance) CreateSnapshot(ctx context.Context) (*Snapshot, error) {
	resp, err := i.c.CreateSnapshotWithResponse(apiv2.WithZone(ctx, i.zone), i.ID)
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return i.c.GetSnapshot(ctx, i.zone, *res.(*papi.Reference).Id)
}

// DetachElasticIP detaches the Compute instance from the specified Elastic IP.
func (i *Instance) DetachElasticIP(ctx context.Context, elasticIP *ElasticIP) error {
	resp, err := i.c.DetachInstanceFromElasticIpWithResponse(
		apiv2.WithZone(ctx, i.zone), elasticIP.ID, papi.DetachInstanceFromElasticIpJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DetachPrivateNetwork detaches the Compute instance from the specified Private Network.
func (i *Instance) DetachPrivateNetwork(ctx context.Context, privateNetwork *PrivateNetwork) error {
	resp, err := i.c.DetachInstanceFromPrivateNetworkWithResponse(
		apiv2.WithZone(ctx, i.zone), privateNetwork.ID, papi.DetachInstanceFromPrivateNetworkJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DetachSecurityGroup detaches the Compute instance from the specified Security Group.
func (i *Instance) DetachSecurityGroup(ctx context.Context, securityGroup *SecurityGroup) error {
	resp, err := i.c.DetachInstanceFromSecurityGroupWithResponse(
		apiv2.WithZone(ctx, i.zone), securityGroup.ID, papi.DetachInstanceFromSecurityGroupJSONRequestBody{
			Instance: papi.Instance{Id: &i.ID},
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
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

// ResetField resets the specified Compute instance field to its default value.
// The value expected for the field parameter is a pointer to the Instance field to reset.
func (i *Instance) ResetField(ctx context.Context, field interface{}) error {
	resetField, err := resetFieldName(i, field)
	if err != nil {
		return err
	}

	resp, err := i.c.ResetInstanceFieldWithResponse(
		apiv2.WithZone(ctx, i.zone),
		i.ID,
		papi.ResetInstanceFieldParamsField(resetField),
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
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
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// SecurityGroups returns the list of Security Groups attached to the Compute instance.
func (i *Instance) SecurityGroups(ctx context.Context) ([]*SecurityGroup, error) {
	res, err := i.c.fetchFromIDs(ctx, i.zone, i.SecurityGroupIDs, new(SecurityGroup))
	return res.([]*SecurityGroup), err
}

// Start starts the Compute instance.
func (i *Instance) Start(ctx context.Context) error {
	resp, err := i.c.StartInstanceWithResponse(apiv2.WithZone(ctx, i.zone), i.ID)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
		Poll(ctx, i.c.OperationPoller(i.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// Stop stops the Compute instance.
func (i *Instance) Stop(ctx context.Context) error {
	resp, err := i.c.StopInstanceWithResponse(apiv2.WithZone(ctx, i.zone), i.ID)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(i.c.timeout).
		WithInterval(i.c.pollInterval).
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
			Labels: func() *papi.Labels {
				if len(instance.Labels) > 0 {
					return &papi.Labels{AdditionalProperties: instance.Labels}
				}
				return nil
			}(),
			Name: &instance.Name,
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
		WithInterval(c.pollInterval).
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

// FindInstance attempts to find a Compute instance by name or ID in the specified zone.
// In case the identifier is a name and multiple resources match, an ErrTooManyFound error is returned.
func (c *Client) FindInstance(ctx context.Context, zone, v string) (*Instance, error) {
	res, err := c.ListInstances(ctx, zone)
	if err != nil {
		return nil, err
	}

	var found *Instance
	for _, r := range res {
		if r.ID == v {
			return c.GetInstance(ctx, zone, r.ID)
		}

		// Historically, the Exoscale API allowed users to create multiple Compute instances sharing a common name.
		// This function being expected to return one resource at most, in case the specified identifier is a name
		// we have to check that there aren't more that one matching result before returning it.
		if r.Name == v {
			if found != nil {
				return nil, apiv2.ErrTooManyFound
			}
			found = r
		}
	}

	if found != nil {
		return found, nil
	}

	return nil, apiv2.ErrNotFound
}

// UpdateInstance updates the specified Compute instance in the specified zone.
func (c *Client) UpdateInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.UpdateInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		instance.ID,
		papi.UpdateInstanceJSONRequestBody{
			Labels: func() *papi.Labels {
				if len(instance.Labels) > 0 {
					return &papi.Labels{AdditionalProperties: instance.Labels}
				}
				return nil
			}(),
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
		WithInterval(c.pollInterval).
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
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
