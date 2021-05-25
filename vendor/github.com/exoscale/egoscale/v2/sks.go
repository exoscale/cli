package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// SKSNodepool represents an SKS Nodepool.
type SKSNodepool struct {
	AntiAffinityGroupIDs []string `reset:"anti-affinity-groups"`
	CreatedAt            time.Time
	DeployTargetID       string `reset:"deploy-target"`
	Description          string `reset:"description"`
	DiskSize             int64
	ID                   string
	InstancePoolID       string
	InstancePrefix       string
	InstanceTypeID       string
	Name                 string
	SecurityGroupIDs     []string `reset:"security-groups"`
	Size                 int64
	State                string
	TemplateID           string
	Version              string

	c    *Client
	zone string
}

func sksNodepoolFromAPI(client *Client, zone string, np *papi.SksNodepool) *SKSNodepool {
	return &SKSNodepool{
		AntiAffinityGroupIDs: func() []string {
			ids := make([]string, 0)
			if np.AntiAffinityGroups != nil {
				for _, aag := range *np.AntiAffinityGroups {
					aag := aag
					ids = append(ids, *aag.Id)
				}
			}
			return ids
		}(),
		CreatedAt: *np.CreatedAt,
		DeployTargetID: func() string {
			if np.DeployTarget != nil {
				return papi.OptionalString(np.DeployTarget.Id)
			}
			return ""
		}(),
		Description:    papi.OptionalString(np.Description),
		DiskSize:       *np.DiskSize,
		ID:             *np.Id,
		InstancePoolID: *np.InstancePool.Id,
		InstancePrefix: papi.OptionalString(np.InstancePrefix),
		InstanceTypeID: *np.InstanceType.Id,
		Name:           *np.Name,
		SecurityGroupIDs: func() []string {
			ids := make([]string, 0)
			if np.SecurityGroups != nil {
				for _, sg := range *np.SecurityGroups {
					sg := sg
					ids = append(ids, *sg.Id)
				}
			}
			return ids
		}(),
		Size:       *np.Size,
		State:      string(*np.State),
		TemplateID: *np.Template.Id,
		Version:    *np.Version,

		c:    client,
		zone: zone,
	}
}

// AntiAffinityGroups returns the list of Anti-Affinity Groups applied to the members of the cluster Nodepool.
func (n *SKSNodepool) AntiAffinityGroups(ctx context.Context) ([]*AntiAffinityGroup, error) {
	res, err := n.c.fetchFromIDs(ctx, n.zone, n.AntiAffinityGroupIDs, new(AntiAffinityGroup))
	return res.([]*AntiAffinityGroup), err
}

// SecurityGroups returns the list of Security Groups attached to the members of the cluster Nodepool.
func (n *SKSNodepool) SecurityGroups(ctx context.Context) ([]*SecurityGroup, error) {
	res, err := n.c.fetchFromIDs(ctx, n.zone, n.SecurityGroupIDs, new(SecurityGroup))
	return res.([]*SecurityGroup), err
}

// SKSCluster represents an SKS cluster.
type SKSCluster struct {
	AddOns       []string
	CNI          string
	CreatedAt    time.Time
	Description  string `reset:"description"`
	Endpoint     string
	ID           string
	Name         string
	Nodepools    []*SKSNodepool
	ServiceLevel string
	State        string
	Version      string

	c    *Client
	zone string
}

func sksClusterFromAPI(client *Client, zone string, c *papi.SksCluster) *SKSCluster {
	return &SKSCluster{
		AddOns: func() []string {
			addOns := make([]string, 0)
			if c.Addons != nil {
				for _, a := range *c.Addons {
					addOns = append(addOns, string(a))
				}
			}
			return addOns
		}(),
		CNI: func() string {
			if c.Cni != nil {
				return string(*c.Cni)
			}
			return ""
		}(),
		CreatedAt:   *c.CreatedAt,
		Description: papi.OptionalString(c.Description),
		Endpoint:    *c.Endpoint,
		ID:          *c.Id,
		Name:        *c.Name,
		Nodepools: func() []*SKSNodepool {
			nodepools := make([]*SKSNodepool, 0)
			if c.Nodepools != nil {
				for _, n := range *c.Nodepools {
					n := n
					nodepools = append(nodepools, sksNodepoolFromAPI(client, zone, &n))
				}
			}
			return nodepools
		}(),
		ServiceLevel: string(*c.Level),
		State:        string(*c.State),
		Version:      *c.Version,

		c:    client,
		zone: zone,
	}
}

// RotateCCMCredentials rotates the Exoscale IAM credentials managed by the SKS control plane for the
// Kubernetes Exoscale Cloud Controller Manager.
func (c *SKSCluster) RotateCCMCredentials(ctx context.Context) error {
	resp, err := c.c.RotateSksCcmCredentialsWithResponse(apiv2.WithZone(ctx, c.zone), c.ID)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// AuthorityCert returns the SKS cluster base64-encoded certificate content for the specified authority.
func (c *SKSCluster) AuthorityCert(ctx context.Context, authority string) (string, error) {
	if authority == "" {
		return "", errors.New("authority not specified")
	}

	resp, err := c.c.GetSksClusterAuthorityCertWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.GetSksClusterAuthorityCertParamsAuthority(authority),
	)
	if err != nil {
		return "", err
	}

	return papi.OptionalString(resp.JSON200.Cacert), nil
}

// RequestKubeconfig returns a base64-encoded kubeconfig content for the specified user name,
// optionally associated to specified groups for a duration d (default API-set TTL applies if not specified).
// Fore more information: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
func (c *SKSCluster) RequestKubeconfig(
	ctx context.Context,
	user string,
	groups []string,
	d time.Duration,
) (string, error) {
	if user == "" {
		return "", errors.New("user not specified")
	}

	resp, err := c.c.GenerateSksClusterKubeconfigWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.GenerateSksClusterKubeconfigJSONRequestBody{
			User:   &user,
			Groups: &groups,
			Ttl: func() *int64 {
				ttl := int64(d.Seconds())
				if ttl > 0 {
					return &ttl
				}
				return nil
			}(),
		})
	if err != nil {
		return "", err
	}

	return papi.OptionalString(resp.JSON200.Kubeconfig), nil
}

// AddNodepool adds a Nodepool to the SKS cluster.
func (c *SKSCluster) AddNodepool(ctx context.Context, np *SKSNodepool) (*SKSNodepool, error) {
	resp, err := c.c.CreateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.CreateSksNodepoolJSONRequestBody{
			AntiAffinityGroups: func() *[]papi.AntiAffinityGroup {
				if l := len(np.AntiAffinityGroupIDs); l > 0 {
					list := make([]papi.AntiAffinityGroup, l)
					for i, v := range np.AntiAffinityGroupIDs {
						v := v
						list[i] = papi.AntiAffinityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
			DeployTarget: func() *papi.DeployTarget {
				if np.DeployTargetID != "" {
					return &papi.DeployTarget{Id: &np.DeployTargetID}
				}
				return nil
			}(),
			Description: func() *string {
				if np.Description != "" {
					return &np.Description
				}
				return nil
			}(),
			DiskSize:       np.DiskSize,
			InstancePrefix: &np.InstancePrefix,
			InstanceType:   papi.InstanceType{Id: &np.InstanceTypeID},
			Name:           np.Name,
			SecurityGroups: func() *[]papi.SecurityGroup {
				if l := len(np.SecurityGroupIDs); l > 0 {
					list := make([]papi.SecurityGroup, l)
					for i, v := range np.SecurityGroupIDs {
						v := v
						list[i] = papi.SecurityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
			Size: np.Size,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nodepoolRes, err := c.c.GetSksNodepoolWithResponse(ctx, c.ID, *res.(*papi.Reference).Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Nodepool: %s", err)
	}

	return sksNodepoolFromAPI(c.c, c.zone, nodepoolRes.JSON200), nil
}

// UpdateNodepool updates the specified SKS cluster Nodepool.
func (c *SKSCluster) UpdateNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.UpdateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.UpdateSksNodepoolJSONRequestBody{
			AntiAffinityGroups: func() *[]papi.AntiAffinityGroup {
				if l := len(np.AntiAffinityGroupIDs); l > 0 {
					list := make([]papi.AntiAffinityGroup, l)
					for i, v := range np.AntiAffinityGroupIDs {
						v := v
						list[i] = papi.AntiAffinityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
			DeployTarget: func() *papi.DeployTarget {
				if np.DeployTargetID != "" {
					return &papi.DeployTarget{Id: &np.DeployTargetID}
				}
				return nil
			}(),
			Description: func() *string {
				if np.Description != "" {
					return &np.Description
				}
				return nil
			}(),
			DiskSize: func() *int64 {
				if np.DiskSize > 0 {
					return &np.DiskSize
				}
				return nil
			}(),
			InstancePrefix: func() *string {
				if np.InstancePrefix != "" {
					return &np.InstancePrefix
				}
				return nil
			}(),
			InstanceType: func() *papi.InstanceType {
				if np.InstanceTypeID != "" {
					return &papi.InstanceType{Id: &np.InstanceTypeID}
				}
				return nil
			}(),
			Name: func() *string {
				if np.Name != "" {
					return &np.Name
				}
				return nil
			}(),
			SecurityGroups: func() *[]papi.SecurityGroup {
				if l := len(np.SecurityGroupIDs); l > 0 {
					list := make([]papi.SecurityGroup, l)
					for i, v := range np.SecurityGroupIDs {
						v := v
						list[i] = papi.SecurityGroup{Id: &v}
					}
					return &list
				}
				return nil
			}(),
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ScaleNodepool scales the SKS cluster Nodepool to the specified number of Kubernetes Nodes.
func (c *SKSCluster) ScaleNodepool(ctx context.Context, np *SKSNodepool, nodes int64) error {
	resp, err := c.c.ScaleSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.ScaleSksNodepoolJSONRequestBody{Size: nodes},
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// EvictNodepoolMembers evicts the specified members (identified by their Compute instance ID) from the
// SKS cluster Nodepool.
func (c *SKSCluster) EvictNodepoolMembers(ctx context.Context, np *SKSNodepool, members []string) error {
	resp, err := c.c.EvictSksNodepoolMembersWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.EvictSksNodepoolMembersJSONRequestBody{Instances: &members},
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteNodepool deletes the specified Nodepool from the SKS cluster.
func (c *SKSCluster) DeleteNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.DeleteSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ResetField resets the specified SKS cluster field to its default value.
// The value expected for the field parameter is a pointer to the SKSCluster field to reset.
func (c *SKSCluster) ResetField(ctx context.Context, field interface{}) error {
	resetField, err := resetFieldName(c, field)
	if err != nil {
		return err
	}

	resp, err := c.c.ResetSksClusterFieldWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.ResetSksClusterFieldParamsField(resetField),
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ResetNodepoolField resets the specified SKS Nodepool field to its default value.
// The value expected for the field parameter is a pointer to the SKSNodepool field to reset.
func (c *SKSCluster) ResetNodepoolField(ctx context.Context, np *SKSNodepool, field interface{}) error {
	resetField, err := resetFieldName(np, field)
	if err != nil {
		return err
	}

	resp, err := c.c.ResetSksNodepoolFieldWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.ResetSksNodepoolFieldParamsField(resetField),
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		WithInterval(c.c.pollInterval).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateSKSCluster creates an SKS cluster in the specified zone.
func (c *Client) CreateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) (*SKSCluster, error) {
	resp, err := c.CreateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateSksClusterJSONRequestBody{
			Addons: func() *[]papi.CreateSksClusterJSONBodyAddons {
				var addOns []papi.CreateSksClusterJSONBodyAddons
				if l := len(cluster.AddOns); l > 0 {
					addOns = make([]papi.CreateSksClusterJSONBodyAddons, l)
					for i := range cluster.AddOns {
						addOns[i] = papi.CreateSksClusterJSONBodyAddons(cluster.AddOns[i])
					}
					return &addOns
				}
				return nil
			}(),
			Cni: func() *papi.CreateSksClusterJSONBodyCni {
				if cluster.CNI != "" {
					return (*papi.CreateSksClusterJSONBodyCni)(&cluster.CNI)
				}
				return nil
			}(),
			Description: &cluster.Description,
			Level:       papi.CreateSksClusterJSONBodyLevel(cluster.ServiceLevel),
			Name:        cluster.Name,
			Version:     cluster.Version,
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

	return c.GetSKSCluster(ctx, zone, *res.(*papi.Reference).Id)
}

// ListSKSClusters returns the list of existing SKS clusters in the specified zone.
func (c *Client) ListSKSClusters(ctx context.Context, zone string) ([]*SKSCluster, error) {
	list := make([]*SKSCluster, 0)

	resp, err := c.ListSksClustersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SksClusters != nil {
		for i := range *resp.JSON200.SksClusters {
			list = append(list, sksClusterFromAPI(c, zone, &(*resp.JSON200.SksClusters)[i]))
		}
	}

	return list, nil
}

// ListSKSClusterVersions returns the list of Kubernetes versions supported during SKS cluster creation.
func (c *Client) ListSKSClusterVersions(ctx context.Context) ([]string, error) {
	list := make([]string, 0)

	resp, err := c.ListSksClusterVersionsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SksClusterVersions != nil {
		for i := range *resp.JSON200.SksClusterVersions {
			version := &(*resp.JSON200.SksClusterVersions)[i]
			list = append(list, *version)
		}
	}

	return list, nil
}

// GetSKSCluster returns the SKS cluster corresponding to the specified ID in the specified zone.
func (c *Client) GetSKSCluster(ctx context.Context, zone, id string) (*SKSCluster, error) {
	resp, err := c.GetSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return sksClusterFromAPI(c, zone, resp.JSON200), nil
}

// FindSKSCluster attempts to find an SKS cluster by name or ID in the specified zone.
func (c *Client) FindSKSCluster(ctx context.Context, zone, v string) (*SKSCluster, error) {
	res, err := c.ListSKSClusters(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if r.ID == v || r.Name == v {
			return c.GetSKSCluster(ctx, zone, r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// UpdateSKSCluster updates the specified SKS cluster in the specified zone.
func (c *Client) UpdateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) error {
	resp, err := c.UpdateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		cluster.ID,
		papi.UpdateSksClusterJSONRequestBody{
			Description: func() *string {
				if cluster.Description != "" {
					return &cluster.Description
				}
				return nil
			}(),
			Name: func() *string {
				if cluster.Name != "" {
					return &cluster.Name
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

// UpgradeSKSCluster upgrades the SKS cluster corresponding to the specified ID in the specified zone to the
// requested Kubernetes version.
func (c *Client) UpgradeSKSCluster(ctx context.Context, zone, id, version string) error {
	resp, err := c.UpgradeSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		id,
		papi.UpgradeSksClusterJSONRequestBody{Version: version})
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

// DeleteSKSCluster deletes the specified SKS cluster in the specified zone.
func (c *Client) DeleteSKSCluster(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
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
