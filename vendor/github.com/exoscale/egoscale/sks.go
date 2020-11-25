package egoscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	v2 "github.com/exoscale/egoscale/internal/v2"
)

// SKSNodepool represents a SKS Nodepool.
type SKSNodepool struct {
	ID               string
	Name             string
	Description      string
	CreatedAt        time.Time
	InstancePoolID   string
	InstanceTypeID   string
	TemplateID       string
	DiskSize         int64
	SecurityGroupIDs []string
	Version          string
	Size             int64
	State            string
}

func sksNodepoolFromAPI(n *v2.SksNodepool) *SKSNodepool {
	return &SKSNodepool{
		ID:             optionalString(n.Id),
		Name:           optionalString(n.Name),
		Description:    optionalString(n.Description),
		CreatedAt:      *n.CreatedAt,
		InstancePoolID: optionalString(n.InstancePool.Id),
		InstanceTypeID: optionalString(n.InstanceType.Id),
		TemplateID:     optionalString(n.Template.Id),
		DiskSize:       optionalInt64(n.DiskSize),
		Version:        optionalString(n.Version),
		Size:           optionalInt64(n.Size),
		State:          optionalString(n.State),
	}
}

// SKSCluster represents a SKS cluster.
type SKSCluster struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	Endpoint    string
	Nodepools   []*SKSNodepool
	Version     string
	State       string

	c    *Client
	zone string
}

func sksClusterFromAPI(c *v2.SksCluster) *SKSCluster {
	return &SKSCluster{
		ID:          optionalString(c.Id),
		Name:        optionalString(c.Name),
		Description: optionalString(c.Description),
		CreatedAt:   *c.CreatedAt,
		Endpoint:    optionalString(c.Endpoint),
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
		Version: optionalString(c.Version),
		State:   optionalString(c.State),
	}
}

// RequestKubeconfig returns a base64-encoded kubeconfig content for the specified user name,
// optionally associated to specified groups for a duration d (default API-set TTL applies if not specified).
// Fore more information: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
func (c *SKSCluster) RequestKubeconfig(user string, groups []string, d time.Duration) (string, error) {
	if user == "" {
		return "", errors.New("user not specified")
	}

	resp, err := c.c.v2.GenerateSksClusterKubeconfigWithResponse(
		apiv2.WithZone(context.Background(), c.zone),
		c.ID,
		v2.GenerateSksClusterKubeconfigJSONRequestBody{
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
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	return optionalString(resp.JSON200.Kubeconfig), nil
}

// AddNodepool adds a Nodepool to the SKS cluster.
func (c *SKSCluster) AddNodepool(ctx context.Context, np *SKSNodepool) (*SKSNodepool, error) {
	resp, err := c.c.v2.CreateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		v2.CreateSksNodepoolJSONRequestBody{
			Description:  &np.Description,
			DiskSize:     &np.DiskSize,
			InstanceType: &v2.InstanceType{Id: &np.InstanceTypeID},
			Name:         &np.Name,
			SecurityGroups: func() *[]v2.SecurityGroup {
				sgs := make([]v2.SecurityGroup, len(np.SecurityGroupIDs))
				for i, sgID := range np.SecurityGroupIDs {
					sgID := sgID
					sgs[i] = v2.SecurityGroup{Id: &sgID}
				}
				return &sgs
			}(),
			Size:    &np.Size,
			Version: &np.Version,
		})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	res, err := v2.NewPoller().
		WithTimeout(c.c.Timeout).
		Poll(ctx, c.c.v2.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nodepoolRes, err := c.c.v2.GetSksNodepoolWithResponse(ctx, c.ID, *res.(*v2.Reference).Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Nodepool: %s", err)
	}

	return sksNodepoolFromAPI(nodepoolRes.JSON200), nil
}

// UpdateNodepool updates the specified SKS cluster Nodepool.
func (c *SKSCluster) UpdateNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.v2.UpdateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		v2.UpdateSksNodepoolJSONRequestBody{
			Name:         &np.Name,
			Description:  &np.Description,
			InstanceType: &v2.InstanceType{Id: &np.InstanceTypeID},
			DiskSize:     &np.DiskSize,
		})
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	_, err = v2.NewPoller().
		WithTimeout(c.c.Timeout).
		Poll(ctx, c.c.v2.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ScaleNodepool scales the SKS cluster Nodepool to the specified number of Kubernetes Nodes.
func (c *SKSCluster) ScaleNodepool(ctx context.Context, np *SKSNodepool, nodes int64) error {
	resp, err := c.c.v2.ScaleSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		v2.ScaleSksNodepoolJSONRequestBody{Size: &nodes},
	)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	_, err = v2.NewPoller().
		WithTimeout(c.c.Timeout).
		Poll(ctx, c.c.v2.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// EvictNodepoolMembers evicts the specified members (identified by their Compute instance ID) from the
// SKS cluster Nodepool.
func (c *SKSCluster) EvictNodepoolMembers(ctx context.Context, np *SKSNodepool, members []string) error {
	instances := make(v2.EvictSksNodepoolMembersJSONRequestBody, len(members))

	for i := range members {
		id := members[i]
		instances[i] = v2.Instance{Id: &id}
	}

	resp, err := c.c.v2.EvictSksNodepoolMembersWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		instances,
	)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	_, err = v2.NewPoller().
		WithTimeout(c.c.Timeout).
		Poll(ctx, c.c.v2.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteNodepool deletes the specified Nodepool from the SKS cluster.
func (c *SKSCluster) DeleteNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.v2.DeleteSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
	)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	_, err = v2.NewPoller().
		WithTimeout(c.c.Timeout).
		Poll(ctx, c.c.v2.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateSKSCluster creates a SKS cluster in the specified zone.
func (c *Client) CreateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) (*SKSCluster, error) {
	resp, err := c.v2.CreateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		v2.CreateSksClusterJSONRequestBody{
			Name:        &cluster.Name,
			Description: &cluster.Description,
			Version:     &cluster.Version,
		})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	res, err := v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetSKSCluster(ctx, zone, *res.(*v2.Reference).Id)
}

// ListSKSClusters returns the list of existing SKS clusters in the specified zone.
func (c *Client) ListSKSClusters(ctx context.Context, zone string) ([]*SKSCluster, error) {
	list := make([]*SKSCluster, 0)

	resp, err := c.v2.ListSksClustersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
	}

	if resp.JSON200.SksClusters != nil {
		for i := range *resp.JSON200.SksClusters {
			cluster := sksClusterFromAPI(&(*resp.JSON200.SksClusters)[i])
			cluster.c = c
			cluster.zone = zone

			list = append(list, cluster)
		}
	}

	return list, nil
}

// GetSKSCluster returns the SKS cluster corresponding to the specified ID in the specified zone.
func (c *Client) GetSKSCluster(ctx context.Context, zone, id string) (*SKSCluster, error) {
	resp, err := c.v2.GetSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return nil, ErrNotFound

		default:
			return nil, fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	cluster := sksClusterFromAPI(resp.JSON200)
	cluster.c = c
	cluster.zone = zone

	return cluster, nil
}

// UpdateSKSCluster updates the specified SKS cluster in the specified zone.
func (c *Client) UpdateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) error {
	resp, err := c.v2.UpdateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		cluster.ID,
		v2.UpdateSksClusterJSONRequestBody{
			Name:        &cluster.Name,
			Description: &cluster.Description,
		})
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return ErrNotFound

		default:
			return fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	_, err = v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteSKSCluster deletes the specified SKS cluster in the specified zone.
func (c *Client) DeleteSKSCluster(ctx context.Context, zone, id string) error {
	resp, err := c.v2.DeleteSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return ErrNotFound

		default:
			return fmt.Errorf("unexpected response from API: %s", resp.Status())
		}
	}

	_, err = v2.NewPoller().
		WithTimeout(c.Timeout).
		Poll(ctx, c.v2.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
