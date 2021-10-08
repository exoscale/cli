package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

// ListSKSClusterVersionsOpt represents an ListSKSClusterVersions operation option.
type ListSKSClusterVersionsOpt func(params *oapi.ListSksClusterVersionsParams)

// ListSKSClusterVersionsWithDeprecated includes deprecated results when listing SKS Cluster versions
// nolint:gocritic
func ListSKSClusterVersionsWithDeprecated(v bool) ListSKSClusterVersionsOpt {
	return func(p *oapi.ListSksClusterVersionsParams) {
		if v {
			vs := "true"
			p.IncludeDeprecated = &vs
		}
	}
}

// SKSNodepoolTaint represents an SKS Nodepool Kubernetes Node taint.
type SKSNodepoolTaint struct {
	Effect string
	Value  string
}

func sksNodepoolTaintFromAPI(t *oapi.SksNodepoolTaint) *SKSNodepoolTaint {
	return &SKSNodepoolTaint{
		Effect: string(t.Effect),
		Value:  t.Value,
	}
}

// SKSNodepool represents an SKS Nodepool.
type SKSNodepool struct {
	AddOns               *[]string
	AntiAffinityGroupIDs *[]string
	CreatedAt            *time.Time
	DeployTargetID       *string
	Description          *string
	DiskSize             *int64  `req-for:"create"`
	ID                   *string `req-for:"update,delete"`
	InstancePoolID       *string
	InstancePrefix       *string
	InstanceTypeID       *string `req-for:"create"`
	Labels               *map[string]string
	Name                 *string `req-for:"create"`
	PrivateNetworkIDs    *[]string
	SecurityGroupIDs     *[]string
	Size                 *int64 `req-for:"create"`
	State                *string
	Taints               *map[string]*SKSNodepoolTaint
	TemplateID           *string
	Version              *string
}

func sksNodepoolFromAPI(n *oapi.SksNodepool) *SKSNodepool {
	return &SKSNodepool{
		AddOns: func() (v *[]string) {
			if n.Addons != nil {
				addOns := make([]string, 0)
				for _, a := range *n.Addons {
					addOns = append(addOns, string(a))
				}
				v = &addOns
			}
			return
		}(),
		AntiAffinityGroupIDs: func() (v *[]string) {
			if n.AntiAffinityGroups != nil && len(*n.AntiAffinityGroups) > 0 {
				ids := make([]string, 0)
				for _, item := range *n.AntiAffinityGroups {
					item := item
					ids = append(ids, *item.Id)
				}
				v = &ids
			}
			return
		}(),
		CreatedAt: n.CreatedAt,
		DeployTargetID: func() (v *string) {
			if n.DeployTarget != nil {
				v = n.DeployTarget.Id
			}
			return
		}(),
		Description:    n.Description,
		DiskSize:       n.DiskSize,
		ID:             n.Id,
		InstancePoolID: n.InstancePool.Id,
		InstancePrefix: n.InstancePrefix,
		InstanceTypeID: n.InstanceType.Id,
		Labels: func() (v *map[string]string) {
			if n.Labels != nil && len(n.Labels.AdditionalProperties) > 0 {
				v = &n.Labels.AdditionalProperties
			}
			return
		}(),
		Name: n.Name,
		PrivateNetworkIDs: func() (v *[]string) {
			if n.PrivateNetworks != nil && len(*n.PrivateNetworks) > 0 {
				ids := make([]string, 0)
				for _, item := range *n.PrivateNetworks {
					item := item
					ids = append(ids, *item.Id)
				}
				v = &ids
			}
			return
		}(),
		SecurityGroupIDs: func() (v *[]string) {
			if n.SecurityGroups != nil && len(*n.SecurityGroups) > 0 {
				ids := make([]string, 0)
				for _, item := range *n.SecurityGroups {
					item := item
					ids = append(ids, *item.Id)
				}
				v = &ids
			}
			return
		}(),
		Size:  n.Size,
		State: (*string)(n.State),
		Taints: func() (v *map[string]*SKSNodepoolTaint) {
			if n.Taints != nil && len(n.Taints.AdditionalProperties) > 0 {
				taints := make(map[string]*SKSNodepoolTaint)
				for k, t := range n.Taints.AdditionalProperties {
					taints[k] = sksNodepoolTaintFromAPI(&t)
				}
				v = &taints
			}
			return
		}(),
		TemplateID: n.Template.Id,
		Version:    n.Version,
	}
}

// SKSCluster represents an SKS cluster.
type SKSCluster struct {
	AddOns       *[]string
	AutoUpgrade  *bool
	CNI          *string
	CreatedAt    *time.Time
	Description  *string
	Endpoint     *string
	ID           *string `req-for:"update,delete"`
	Labels       *map[string]string
	Name         *string `req-for:"create"`
	Nodepools    []*SKSNodepool
	ServiceLevel *string `req-for:"create"`
	State        *string
	Version      *string `req-for:"create"`
	Zone         *string
}

func sksClusterFromAPI(c *oapi.SksCluster, zone string) *SKSCluster {
	return &SKSCluster{
		AddOns: func() (v *[]string) {
			if c.Addons != nil {
				addOns := make([]string, 0)
				for _, a := range *c.Addons {
					addOns = append(addOns, string(a))
				}
				v = &addOns
			}
			return
		}(),
		AutoUpgrade: c.AutoUpgrade,
		CNI:         (*string)(c.Cni),
		CreatedAt:   c.CreatedAt,
		Description: c.Description,
		Endpoint:    c.Endpoint,
		ID:          c.Id,
		Labels: func() (v *map[string]string) {
			if c.Labels != nil && len(c.Labels.AdditionalProperties) > 0 {
				v = &c.Labels.AdditionalProperties
			}
			return
		}(),
		Name: c.Name,
		Nodepools: func() []*SKSNodepool {
			nodepools := make([]*SKSNodepool, 0)
			if c.Nodepools != nil {
				for _, n := range *c.Nodepools {
					n := n
					nodepools = append(nodepools, sksNodepoolFromAPI(&n))
				}
			}
			return nodepools
		}(),
		ServiceLevel: (*string)(c.Level),
		State:        (*string)(c.State),
		Version:      c.Version,
		Zone:         &zone,
	}
}

// CreateSKSCluster creates an SKS cluster.
func (c *Client) CreateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) (*SKSCluster, error) {
	if err := validateOperationParams(cluster, "create"); err != nil {
		return nil, err
	}

	resp, err := c.CreateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		oapi.CreateSksClusterJSONRequestBody{
			Addons: func() (v *[]oapi.CreateSksClusterJSONBodyAddons) {
				if cluster.AddOns != nil {
					addOns := make([]oapi.CreateSksClusterJSONBodyAddons, len(*cluster.AddOns))
					for i, a := range *cluster.AddOns {
						addOns[i] = oapi.CreateSksClusterJSONBodyAddons(a)
					}
					v = &addOns
				}
				return
			}(),
			AutoUpgrade: cluster.AutoUpgrade,
			Cni:         (*oapi.CreateSksClusterJSONBodyCni)(cluster.CNI),
			Description: cluster.Description,
			Labels: func() (v *oapi.Labels) {
				if cluster.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *cluster.Labels}
				}
				return
			}(),
			Level:   oapi.CreateSksClusterJSONBodyLevel(*cluster.ServiceLevel),
			Name:    *cluster.Name,
			Version: *cluster.Version,
		})
	if err != nil {
		return nil, err
	}

	res, err := oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetSKSCluster(ctx, zone, *res.(*oapi.Reference).Id)
}

// CreateSKSNodepool create an SKS Nodepool.
func (c *Client) CreateSKSNodepool(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	nodepool *SKSNodepool,
) (*SKSNodepool, error) {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return nil, err
	}
	if err := validateOperationParams(nodepool, "create"); err != nil {
		return nil, err
	}

	resp, err := c.CreateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		oapi.CreateSksNodepoolJSONRequestBody{
			Addons: func() (v *[]oapi.CreateSksNodepoolJSONBodyAddons) {
				if nodepool.AddOns != nil {
					addOns := make([]oapi.CreateSksNodepoolJSONBodyAddons, len(*nodepool.AddOns))
					for i, a := range *nodepool.AddOns {
						addOns[i] = oapi.CreateSksNodepoolJSONBodyAddons(a)
					}
					v = &addOns
				}
				return
			}(),
			AntiAffinityGroups: func() (v *[]oapi.AntiAffinityGroup) {
				if nodepool.AntiAffinityGroupIDs != nil {
					ids := make([]oapi.AntiAffinityGroup, len(*nodepool.AntiAffinityGroupIDs))
					for i, item := range *nodepool.AntiAffinityGroupIDs {
						item := item
						ids[i] = oapi.AntiAffinityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			DeployTarget: func() (v *oapi.DeployTarget) {
				if nodepool.DeployTargetID != nil {
					v = &oapi.DeployTarget{Id: nodepool.DeployTargetID}
				}
				return
			}(),
			Description:    nodepool.Description,
			DiskSize:       *nodepool.DiskSize,
			InstancePrefix: nodepool.InstancePrefix,
			InstanceType:   oapi.InstanceType{Id: nodepool.InstanceTypeID},
			Labels: func() (v *oapi.Labels) {
				if nodepool.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *nodepool.Labels}
				}
				return
			}(),
			Name: *nodepool.Name,
			PrivateNetworks: func() (v *[]oapi.PrivateNetwork) {
				if nodepool.PrivateNetworkIDs != nil {
					ids := make([]oapi.PrivateNetwork, len(*nodepool.PrivateNetworkIDs))
					for i, item := range *nodepool.PrivateNetworkIDs {
						item := item
						ids[i] = oapi.PrivateNetwork{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			SecurityGroups: func() (v *[]oapi.SecurityGroup) {
				if nodepool.SecurityGroupIDs != nil {
					ids := make([]oapi.SecurityGroup, len(*nodepool.SecurityGroupIDs))
					for i, item := range *nodepool.SecurityGroupIDs {
						item := item
						ids[i] = oapi.SecurityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			Size: *nodepool.Size,
			Taints: func() (v *oapi.SksNodepoolTaints) {
				if nodepool.Taints != nil {
					taints := oapi.SksNodepoolTaints{AdditionalProperties: map[string]oapi.SksNodepoolTaint{}}
					for k, t := range *nodepool.Taints {
						taints.AdditionalProperties[k] = oapi.SksNodepoolTaint{
							Effect: (oapi.SksNodepoolTaintEffect)(t.Effect),
							Value:  t.Value,
						}
					}
					v = &taints
				}
				return
			}(),
		})
	if err != nil {
		return nil, err
	}

	res, err := oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nodepoolRes, err := c.GetSksNodepoolWithResponse(ctx, *cluster.ID, *res.(*oapi.Reference).Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Nodepool: %s", err)
	}

	return sksNodepoolFromAPI(nodepoolRes.JSON200), nil
}

// DeleteSKSCluster deletes an SKS cluster.
func (c *Client) DeleteSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) error {
	if err := validateOperationParams(cluster, "delete"); err != nil {
		return err
	}

	resp, err := c.DeleteSksClusterWithResponse(apiv2.WithZone(ctx, zone), *cluster.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteSKSNodepool deletes an SKS Nodepool.
func (c *Client) DeleteSKSNodepool(ctx context.Context, zone string, cluster *SKSCluster, nodepool *SKSNodepool) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}
	if err := validateOperationParams(nodepool, "delete"); err != nil {
		return err
	}

	resp, err := c.DeleteSksNodepoolWithResponse(apiv2.WithZone(ctx, zone), *cluster.ID, *nodepool.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// EvictSKSNodepoolMembers evicts the specified members (identified by their Compute instance ID) from the
// SKS cluster Nodepool.
func (c *Client) EvictSKSNodepoolMembers(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	nodepool *SKSNodepool,
	members []string,
) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}
	if err := validateOperationParams(nodepool, "update"); err != nil {
		return err
	}

	resp, err := c.EvictSksNodepoolMembersWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		*nodepool.ID,
		oapi.EvictSksNodepoolMembersJSONRequestBody{Instances: &members},
	)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// FindSKSCluster attempts to find an SKS cluster by name or ID.
func (c *Client) FindSKSCluster(ctx context.Context, zone, x string) (*SKSCluster, error) {
	res, err := c.ListSKSClusters(ctx, zone)
	if err != nil {
		return nil, err
	}

	for _, r := range res {
		if *r.ID == x || *r.Name == x {
			return c.GetSKSCluster(ctx, zone, *r.ID)
		}
	}

	return nil, apiv2.ErrNotFound
}

// GetSKSCluster returns the SKS cluster corresponding to the specified ID.
func (c *Client) GetSKSCluster(ctx context.Context, zone, id string) (*SKSCluster, error) {
	resp, err := c.GetSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	return sksClusterFromAPI(resp.JSON200, zone), nil
}

// GetSKSClusterAuthorityCert returns the SKS cluster base64-encoded certificate content for the specified authority.
func (c *Client) GetSKSClusterAuthorityCert(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	authority string,
) (string, error) {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return "", err
	}

	if authority == "" {
		return "", errors.New("authority not specified")
	}

	resp, err := c.GetSksClusterAuthorityCertWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		oapi.GetSksClusterAuthorityCertParamsAuthority(authority),
	)
	if err != nil {
		return "", err
	}

	return oapi.OptionalString(resp.JSON200.Cacert), nil
}

// GetSKSClusterKubeconfig returns a base64-encoded kubeconfig content for the specified user name, optionally
// associated to specified groups for a duration d (default API-set TTL applies if not specified).
// Fore more information: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
func (c *Client) GetSKSClusterKubeconfig(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	user string,
	groups []string,
	d time.Duration,
) (string, error) {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return "", err
	}

	if user == "" {
		return "", errors.New("user not specified")
	}

	resp, err := c.GenerateSksClusterKubeconfigWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		oapi.GenerateSksClusterKubeconfigJSONRequestBody{
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

	return oapi.OptionalString(resp.JSON200.Kubeconfig), nil
}

// ListSKSClusters returns the list of existing SKS clusters.
func (c *Client) ListSKSClusters(ctx context.Context, zone string) ([]*SKSCluster, error) {
	list := make([]*SKSCluster, 0)

	resp, err := c.ListSksClustersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SksClusters != nil {
		for i := range *resp.JSON200.SksClusters {
			list = append(list, sksClusterFromAPI(&(*resp.JSON200.SksClusters)[i], zone))
		}
	}

	return list, nil
}

// ListSKSClusterVersions returns the list of Kubernetes versions supported during SKS cluster creation.
func (c *Client) ListSKSClusterVersions(ctx context.Context, opts ...ListSKSClusterVersionsOpt) ([]string, error) {
	list := make([]string, 0)

	params := oapi.ListSksClusterVersionsParams{}

	for _, opt := range opts {
		opt(&params)
	}

	resp, err := c.ListSksClusterVersionsWithResponse(ctx, &params)
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

// RotateSKSClusterCCMCredentials rotates the Exoscale IAM credentials managed by the SKS control plane for the
// Kubernetes Exoscale Cloud Controller Manager.
func (c *Client) RotateSKSClusterCCMCredentials(ctx context.Context, zone string, cluster *SKSCluster) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}

	resp, err := c.RotateSksCcmCredentialsWithResponse(apiv2.WithZone(ctx, zone), *cluster.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ScaleSKSNodepool scales the SKS cluster Nodepool to the specified number of Kubernetes Nodes.
func (c *Client) ScaleSKSNodepool(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	nodepool *SKSNodepool,
	size int64,
) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}
	if err := validateOperationParams(nodepool, "update"); err != nil {
		return err
	}

	resp, err := c.ScaleSksNodepoolWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		*nodepool.ID,
		oapi.ScaleSksNodepoolJSONRequestBody{Size: size},
	)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// UpdateSKSCluster updates an SKS cluster.
func (c *Client) UpdateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}

	resp, err := c.UpdateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		oapi.UpdateSksClusterJSONRequestBody{
			AutoUpgrade: cluster.AutoUpgrade,
			Description: cluster.Description,
			Labels: func() (v *oapi.Labels) {
				if cluster.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *cluster.Labels}
				}
				return
			}(),
			Name: cluster.Name,
		})
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// UpdateSKSNodepool updates an SKS Nodepool.
func (c *Client) UpdateSKSNodepool(
	ctx context.Context,
	zone string,
	cluster *SKSCluster,
	nodepool *SKSNodepool,
) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}
	if err := validateOperationParams(nodepool, "update"); err != nil {
		return err
	}

	resp, err := c.UpdateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		*nodepool.ID,
		oapi.UpdateSksNodepoolJSONRequestBody{
			AntiAffinityGroups: func() (v *[]oapi.AntiAffinityGroup) {
				if nodepool.AntiAffinityGroupIDs != nil {
					ids := make([]oapi.AntiAffinityGroup, len(*nodepool.AntiAffinityGroupIDs))
					for i, item := range *nodepool.AntiAffinityGroupIDs {
						item := item
						ids[i] = oapi.AntiAffinityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			DeployTarget: func() (v *oapi.DeployTarget) {
				if nodepool.DeployTargetID != nil {
					v = &oapi.DeployTarget{Id: nodepool.DeployTargetID}
				}
				return
			}(),
			Description:    nodepool.Description,
			DiskSize:       nodepool.DiskSize,
			InstancePrefix: nodepool.InstancePrefix,
			InstanceType: func() (v *oapi.InstanceType) {
				if nodepool.InstanceTypeID != nil {
					v = &oapi.InstanceType{Id: nodepool.InstanceTypeID}
				}
				return
			}(),
			Labels: func() (v *oapi.Labels) {
				if nodepool.Labels != nil {
					v = &oapi.Labels{AdditionalProperties: *nodepool.Labels}
				}
				return
			}(),
			Name: nodepool.Name,
			PrivateNetworks: func() (v *[]oapi.PrivateNetwork) {
				if nodepool.PrivateNetworkIDs != nil {
					ids := make([]oapi.PrivateNetwork, len(*nodepool.PrivateNetworkIDs))
					for i, item := range *nodepool.PrivateNetworkIDs {
						item := item
						ids[i] = oapi.PrivateNetwork{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			SecurityGroups: func() (v *[]oapi.SecurityGroup) {
				if nodepool.SecurityGroupIDs != nil {
					ids := make([]oapi.SecurityGroup, len(*nodepool.SecurityGroupIDs))
					for i, item := range *nodepool.SecurityGroupIDs {
						item := item
						ids[i] = oapi.SecurityGroup{Id: &item}
					}
					v = &ids
				}
				return
			}(),
			Taints: func() (v *oapi.SksNodepoolTaints) {
				if nodepool.Taints != nil {
					taints := oapi.SksNodepoolTaints{AdditionalProperties: map[string]oapi.SksNodepoolTaint{}}
					for k, t := range *nodepool.Taints {
						taints.AdditionalProperties[k] = oapi.SksNodepoolTaint{
							Effect: (oapi.SksNodepoolTaintEffect)(t.Effect),
							Value:  t.Value,
						}
					}
					v = &taints
				}
				return
			}(),
		})
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// UpgradeSKSCluster upgrades an SKS cluster to the requested Kubernetes version.
func (c *Client) UpgradeSKSCluster(ctx context.Context, zone string, cluster *SKSCluster, version string) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}

	resp, err := c.UpgradeSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		*cluster.ID,
		oapi.UpgradeSksClusterJSONRequestBody{Version: version})
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// UpgradeSKSClusterServiceLevel upgrades an SKS cluster to service level "pro".
func (c *Client) UpgradeSKSClusterServiceLevel(ctx context.Context, zone string, cluster *SKSCluster) error {
	if err := validateOperationParams(cluster, "update"); err != nil {
		return err
	}

	resp, err := c.UpgradeSksClusterServiceLevelWithResponse(apiv2.WithZone(ctx, zone), *cluster.ID)
	if err != nil {
		return err
	}

	_, err = oapi.NewPoller().
		WithTimeout(c.timeout).
		WithInterval(c.pollInterval).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
