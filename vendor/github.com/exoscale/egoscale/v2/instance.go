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
	AntiAffinityGroupIDs *[]string
	CreatedAt            *time.Time
	DeployTargetID       *string
	DiskSize             *int64 `req-if:"create"`
	ElasticIPIDs         *[]string
	ID                   *string `req-if:"update"`
	IPv6Address          *net.IP
	IPv6Enabled          *bool
	InstanceTypeID       *string `req-if:"create"`
	Labels               *map[string]string
	Manager              *InstanceManager
	Name                 *string `req-if:"create"`
	PrivateNetworkIDs    *[]string
	PublicIPAddress      *net.IP
	SSHKey               *string
	SecurityGroupIDs     *[]string
	SnapshotIDs          *[]string
	State                *string
	TemplateID           *string `req-if:"create"`
	UserData             *string
}

func instanceFromAPI(i *papi.Instance) *Instance {
	return &Instance{
		AntiAffinityGroupIDs: func() (v *[]string) {
			if i.AntiAffinityGroups != nil && len(*i.AntiAffinityGroups) > 0 {
				ids := make([]string, len(*i.AntiAffinityGroups))
				for i, item := range *i.AntiAffinityGroups {
					ids[i] = *item.Id
				}
				v = &ids
			}
			return
		}(),
		CreatedAt: i.CreatedAt,
		DeployTargetID: func() (v *string) {
			if i.DeployTarget != nil {
				v = i.DeployTarget.Id
			}
			return
		}(),
		DiskSize: i.DiskSize,
		ElasticIPIDs: func() (v *[]string) {
			if i.ElasticIps != nil && len(*i.ElasticIps) > 0 {
				ids := make([]string, len(*i.ElasticIps))
				for i, item := range *i.ElasticIps {
					ids[i] = *item.Id
				}
				v = &ids
			}
			return
		}(),
		ID: i.Id,
		IPv6Address: func() (v *net.IP) {
			if i.Ipv6Address != nil {
				ip := net.ParseIP(*i.Ipv6Address)
				v = &ip
			}
			return
		}(),
		IPv6Enabled: func() (v *bool) {
			if i.Ipv6Address != nil {
				ipv6Enabled := i.Ipv6Address != nil
				v = &ipv6Enabled
			}
			return
		}(),
		InstanceTypeID: i.InstanceType.Id,
		Labels: func() (v *map[string]string) {
			if i.Labels != nil && len(i.Labels.AdditionalProperties) > 0 {
				v = &i.Labels.AdditionalProperties
			}
			return
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
		Name: i.Name,
		PrivateNetworkIDs: func() (v *[]string) {
			if i.PrivateNetworks != nil && len(*i.PrivateNetworks) > 0 {
				ids := make([]string, len(*i.PrivateNetworks))
				for i, item := range *i.PrivateNetworks {
					ids[i] = *item.Id
				}
				v = &ids
			}
			return
		}(),
		PublicIPAddress: func() (v *net.IP) {
			if i.PublicIp != nil {
				ip := net.ParseIP(*i.PublicIp)
				v = &ip
			}
			return
		}(),
		SSHKey: func() (v *string) {
			if i.SshKey != nil {
				v = i.SshKey.Name
			}
			return
		}(),
		SecurityGroupIDs: func() (v *[]string) {
			if i.SecurityGroups != nil && len(*i.SecurityGroups) > 0 {
				ids := make([]string, len(*i.SecurityGroups))
				for i, item := range *i.SecurityGroups {
					ids[i] = *item.Id
				}
				v = &ids
			}
			return
		}(),
		SnapshotIDs: func() (v *[]string) {
			if i.Snapshots != nil && len(*i.Snapshots) > 0 {
				ids := make([]string, len(*i.Snapshots))
				for i, item := range *i.Snapshots {
					ids[i] = *item.Id
				}
				v = &ids
			}
			return
		}(),
		State:      (*string)(i.State),
		TemplateID: i.Template.Id,
		UserData:   i.UserData,
	}
}

// ToAPIMock returns the low-level representation of the resource. This is intended for testing purposes.
func (i Instance) ToAPIMock() interface{} {
	return papi.Instance{
		AntiAffinityGroups: func() *[]papi.AntiAffinityGroup {
			if i.AntiAffinityGroupIDs != nil {
				list := make([]papi.AntiAffinityGroup, len(*i.AntiAffinityGroupIDs))
				for j, id := range *i.AntiAffinityGroupIDs {
					id := id
					list[j] = papi.AntiAffinityGroup{Id: &id}
				}
				return &list
			}
			return nil
		}(),
		CreatedAt:    i.CreatedAt,
		DeployTarget: &papi.DeployTarget{Id: i.DeployTargetID},
		DiskSize:     i.DiskSize,
		ElasticIps: func() *[]papi.ElasticIp {
			if i.ElasticIPIDs != nil {
				list := make([]papi.ElasticIp, len(*i.ElasticIPIDs))
				for j, id := range *i.ElasticIPIDs {
					id := id
					list[j] = papi.ElasticIp{Id: &id}
				}
				return &list
			}
			return nil
		}(),
		Id:           i.ID,
		InstanceType: &papi.InstanceType{Id: i.InstanceTypeID},
		Ipv6Address: func() *string {
			if i.IPv6Address != nil {
				v := i.IPv6Address.String()
				return &v
			}
			return nil
		}(),
		Labels: func() *papi.Labels {
			if i.Labels != nil {
				return &papi.Labels{AdditionalProperties: *i.Labels}
			}
			return nil
		}(),
		Manager: func() *papi.Manager {
			if i.Manager != nil {
				return &papi.Manager{
					Id:   &i.Manager.ID,
					Type: (*papi.ManagerType)(&i.Manager.Type),
				}
			}
			return nil
		}(),
		Name: i.Name,
		PrivateNetworks: func() *[]papi.PrivateNetwork {
			if i.PrivateNetworkIDs != nil {
				list := make([]papi.PrivateNetwork, len(*i.PrivateNetworkIDs))
				for j, id := range *i.PrivateNetworkIDs {
					id := id
					list[j] = papi.PrivateNetwork{Id: &id}
				}
				return &list
			}
			return nil
		}(),
		PublicIp: func() *string {
			if i.PublicIPAddress != nil {
				v := i.PublicIPAddress.String()
				return &v
			}
			return nil
		}(),
		SecurityGroups: func() *[]papi.SecurityGroup {
			if i.SecurityGroupIDs != nil {
				list := make([]papi.SecurityGroup, len(*i.SecurityGroupIDs))
				for j, id := range *i.SecurityGroupIDs {
					id := id
					list[j] = papi.SecurityGroup{Id: &id}
				}
				return &list
			}
			return nil
		}(),
		Snapshots: func() *[]papi.Snapshot {
			if i.SnapshotIDs != nil {
				list := make([]papi.Snapshot, len(*i.SnapshotIDs))
				for j, id := range *i.SnapshotIDs {
					id := id
					list[j] = papi.Snapshot{Id: &id}
				}
				return &list
			}
			return nil
		}(),
		SshKey: func() *papi.SshKey {
			if i.SSHKey != nil {
				return &papi.SshKey{Name: i.SSHKey}
			}
			return nil
		}(),
		State:    (*papi.InstanceState)(i.State),
		Template: &papi.Template{Id: i.TemplateID},
		UserData: i.UserData,
	}
}

// AttachInstanceToElasticIP attaches a Compute instance to the specified Elastic IP.
func (c *Client) AttachInstanceToElasticIP(
	ctx context.Context,
	zone string,
	instance *Instance,
	elasticIP *ElasticIP,
) error {
	resp, err := c.AttachInstanceToElasticIpWithResponse(
		apiv2.WithZone(ctx, zone), *elasticIP.ID, papi.AttachInstanceToElasticIpJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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

// AttachInstanceToPrivateNetwork attaches a Compute instance to the specified Private Network.
// If address is specified, it will be used when requesting a network address lease.
func (c *Client) AttachInstanceToPrivateNetwork(
	ctx context.Context,
	zone string,
	instance *Instance,
	privateNetwork *PrivateNetwork,
	address net.IP,
) error {
	resp, err := c.AttachInstanceToPrivateNetworkWithResponse(
		apiv2.WithZone(ctx, zone), *privateNetwork.ID, papi.AttachInstanceToPrivateNetworkJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// AttachInstanceToSecurityGroup attaches a Compute instance to the specified Security Group.
func (c *Client) AttachInstanceToSecurityGroup(
	ctx context.Context,
	zone string,
	instance *Instance,
	securityGroup *SecurityGroup,
) error {
	resp, err := c.AttachInstanceToSecurityGroupWithResponse(
		apiv2.WithZone(ctx, zone), *securityGroup.ID, papi.AttachInstanceToSecurityGroupJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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

// CreateInstance creates a Compute instance.
func (c *Client) CreateInstance(ctx context.Context, zone string, instance *Instance) (*Instance, error) {
	if err := validateOperationParams(instance, "create"); err != nil {
		return nil, err
	}

	resp, err := c.CreateInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateInstanceJSONRequestBody{
			AntiAffinityGroups: func() (v *[]papi.AntiAffinityGroup) {
				if instance.AntiAffinityGroupIDs != nil {
					ids := make([]papi.AntiAffinityGroup, len(*instance.AntiAffinityGroupIDs))
					for i, item := range *instance.AntiAffinityGroupIDs {
						item := item
						ids[i] = papi.AntiAffinityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			DeployTarget: func() (v *papi.DeployTarget) {
				if instance.DeployTargetID != nil {
					v = &papi.DeployTarget{Id: instance.DeployTargetID}
				}
				return
			}(),
			DiskSize:     *instance.DiskSize,
			InstanceType: papi.InstanceType{Id: instance.InstanceTypeID},
			Ipv6Enabled:  instance.IPv6Enabled,
			Labels: func() (v *papi.Labels) {
				if instance.Labels != nil {
					v = &papi.Labels{AdditionalProperties: *instance.Labels}
				}
				return
			}(),
			Name: instance.Name,
			SecurityGroups: func() (v *[]papi.SecurityGroup) {
				if instance.SecurityGroupIDs != nil {
					ids := make([]papi.SecurityGroup, len(*instance.SecurityGroupIDs))
					for i, item := range *instance.SecurityGroupIDs {
						item := item
						ids[i] = papi.SecurityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			SshKey: func() (v *papi.SshKey) {
				if instance.SSHKey != nil {
					v = &papi.SshKey{Name: instance.SSHKey}
				}
				return
			}(),
			Template: papi.Template{Id: instance.TemplateID},
			UserData: instance.UserData,
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

// CreateInstanceSnapshot creates a Snapshot of a Compute instance storage volume.
func (c *Client) CreateInstanceSnapshot(ctx context.Context, zone string, instance *Instance) (*Snapshot, error) {
	resp, err := c.CreateSnapshotWithResponse(apiv2.WithZone(ctx, zone), *instance.ID)
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

	return c.GetSnapshot(ctx, zone, *res.(*papi.Reference).Id)
}

// DetachInstanceFromElasticIP detaches a Compute instance from the specified Elastic IP.
func (c *Client) DetachInstanceFromElasticIP(
	ctx context.Context,
	zone string,
	instance *Instance,
	elasticIP *ElasticIP,
) error {
	resp, err := c.DetachInstanceFromElasticIpWithResponse(
		apiv2.WithZone(ctx, zone), *elasticIP.ID, papi.DetachInstanceFromElasticIpJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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

// DeleteInstance deletes a Compute instance.
func (c *Client) DeleteInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.DeleteInstanceWithResponse(apiv2.WithZone(ctx, zone), *instance.ID)
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

// DetachInstanceFromPrivateNetwork detaches a Compute instance from the specified Private Network.
func (c *Client) DetachInstanceFromPrivateNetwork(
	ctx context.Context,
	zone string,
	instance *Instance,
	privateNetwork *PrivateNetwork,
) error {
	resp, err := c.DetachInstanceFromPrivateNetworkWithResponse(
		apiv2.WithZone(ctx, zone), *privateNetwork.ID, papi.DetachInstanceFromPrivateNetworkJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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

// DetachInstanceFromSecurityGroup detaches a Compute instance from the specified Security Group.
func (c *Client) DetachInstanceFromSecurityGroup(
	ctx context.Context,
	zone string,
	instance *Instance,
	securityGroup *SecurityGroup,
) error {
	resp, err := c.DetachInstanceFromSecurityGroupWithResponse(
		apiv2.WithZone(ctx, zone), *securityGroup.ID, papi.DetachInstanceFromSecurityGroupJSONRequestBody{
			Instance: papi.Instance{Id: instance.ID},
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

// FindInstance attempts to find a Compute instance by name or ID.
// In case the identifier is a name and multiple resources match, an ErrTooManyFound error is returned.
func (c *Client) FindInstance(ctx context.Context, zone, x string) (*Instance, error) {
	res, err := c.ListInstances(ctx, zone)
	if err != nil {
		return nil, err
	}

	var found *Instance
	for _, r := range res {
		if *r.ID == x {
			return c.GetInstance(ctx, zone, *r.ID)
		}

		// Historically, the Exoscale API allowed users to create multiple Compute instances sharing a common name.
		// This function being expected to return one resource at most, in case the specified identifier is a name
		// we have to check that there aren't more than one matching result before returning it.
		if *r.Name == x {
			if found != nil {
				return nil, apiv2.ErrTooManyFound
			}
			found = r
		}
	}

	if found != nil {
		return c.GetInstance(ctx, zone, *found.ID)
	}

	return nil, apiv2.ErrNotFound
}

// GetInstance returns the Instance corresponding to the specified ID.
func (c *Client) GetInstance(ctx context.Context, zone, id string) (*Instance, error) {
	resp, err := c.GetInstanceWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return instanceFromAPI(resp.JSON200), nil
}

// ListInstances returns the list of existing Compute instances.
func (c *Client) ListInstances(ctx context.Context, zone string) ([]*Instance, error) {
	list := make([]*Instance, 0)

	resp, err := c.ListInstancesWithResponse(apiv2.WithZone(ctx, zone), &papi.ListInstancesParams{})
	if err != nil {
		return nil, err
	}

	if resp.JSON200.Instances != nil {
		for i := range *resp.JSON200.Instances {
			list = append(list, instanceFromAPI(&(*resp.JSON200.Instances)[i]))
		}
	}

	return list, nil
}

// RebootInstance reboots a Compute instance.
func (c *Client) RebootInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.RebootInstanceWithResponse(apiv2.WithZone(ctx, zone), *instance.ID)
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

// ResetInstance resets a Compute instance to a base template state (the instance's current template if not specified),
// and optionally resizes its disk size.
func (c *Client) ResetInstance(
	ctx context.Context,
	zone string,
	instance *Instance,
	template *Template,
	diskSize int64,
) error {
	resp, err := c.ResetInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		*instance.ID,
		papi.ResetInstanceJSONRequestBody{
			DiskSize: func() (v *int64) {
				if diskSize > 0 {
					v = &diskSize
				}
				return
			}(),
			Template: func() (v *papi.Template) {
				if template != nil {
					v = &papi.Template{Id: template.ID}
				}
				return
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

// ResizeInstanceDisk resizes a Compute instance's disk to a larger size.
func (c *Client) ResizeInstanceDisk(ctx context.Context, zone string, instance *Instance, size int64) error {
	resp, err := c.ResizeInstanceDiskWithResponse(
		apiv2.WithZone(ctx, zone),
		*instance.ID,
		papi.ResizeInstanceDiskJSONRequestBody{DiskSize: size})
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

// RevertInstanceToSnapshot reverts a Compute instance storage volume to the specified Snapshot.
func (c *Client) RevertInstanceToSnapshot(
	ctx context.Context,
	zone string,
	instance *Instance,
	snapshot *Snapshot,
) error {
	resp, err := c.RevertInstanceToSnapshotWithResponse(
		apiv2.WithZone(ctx, zone),
		*instance.ID,
		papi.RevertInstanceToSnapshotJSONRequestBody{Id: *snapshot.ID})
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

// ScaleInstance scales a Compute instance type.
func (c *Client) ScaleInstance(ctx context.Context, zone string, instance *Instance, instanceType *InstanceType) error {
	resp, err := c.ScaleInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		*instance.ID,
		papi.ScaleInstanceJSONRequestBody{InstanceType: papi.InstanceType{Id: instanceType.ID}},
	)
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

// StartInstance starts a Compute instance.
func (c *Client) StartInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.StartInstanceWithResponse(apiv2.WithZone(ctx, zone), *instance.ID)
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

// StopInstance stops a Compute instance.
func (c *Client) StopInstance(ctx context.Context, zone string, instance *Instance) error {
	resp, err := c.StopInstanceWithResponse(apiv2.WithZone(ctx, zone), *instance.ID)
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

// UpdateInstance updates a Compute instance.
func (c *Client) UpdateInstance(ctx context.Context, zone string, instance *Instance) error {
	if err := validateOperationParams(instance, "update"); err != nil {
		return err
	}

	resp, err := c.UpdateInstanceWithResponse(
		apiv2.WithZone(ctx, zone),
		*instance.ID,
		papi.UpdateInstanceJSONRequestBody{
			Labels: func() (v *papi.Labels) {
				if instance.Labels != nil {
					v = &papi.Labels{AdditionalProperties: *instance.Labels}
				}
				return
			}(),
			Name:     instance.Name,
			UserData: instance.UserData,
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
