package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v3 "github.com/exoscale/egoscale/v3"
)

var sksNodepoolCmd = &cobra.Command{
	Use:     "nodepool",
	Short:   "Manage SKS cluster Nodepools",
	Aliases: []string{"np"},
}
var errExpectedFormatNodepoolTaint = errors.New("expected format KEY=VALUE:EFFECT")

// parseSKSNodepoolTaint parses a CLI-formatted Kubernetes Node taint.
// According to:
// https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#taint
// We will support only: KEY=VALUE:EFFECT for the moment as the API support only this format.
// or an error if the input value parsing failed.

func parseSKSNodepoolTaint(v string) (string, *v3.SKSNodepoolTaint, error) {
	kv := strings.Split(v, "=")
	if len(kv) != 2 {
		return "", nil, errExpectedFormatNodepoolTaint
	}

	valueEffect := strings.Split(kv[1], ":")
	if len(valueEffect) != 2 {
		return "", nil, errExpectedFormatNodepoolTaint
	}

	taintKey := kv[0]
	taintValue := valueEffect[0]
	taintEffect := valueEffect[1]

	if taintKey == "" || taintValue == "" || taintEffect == "" {
		return "", nil, errExpectedFormatNodepoolTaint
	}

	return taintKey, &v3.SKSNodepoolTaint{Effect: v3.SKSNodepoolTaintEffect(taintEffect), Value: taintValue}, nil
}

func createNodepoolRequest(
	ctx context.Context,
	client *v3.Client,
	name string,
	description string,
	diskSize int64,
	instancePrefix string,
	size int64,
	instanceType string,
	labels map[string]string,
	antiAffinityGroups []string,
	deployTarget string,
	privateNetworks []string,
	securityGroups []string,
	taints []string,
	kubeletImageGC *v3.KubeletImageGC,
) (v3.CreateSKSNodepoolRequest, error) {

	nodepoolReq := v3.CreateSKSNodepoolRequest{
		Description:    description,
		DiskSize:       diskSize,
		InstancePrefix: instancePrefix,
		Name:           name,
		Size:           size,
		Labels:         labels,
		KubeletImageGC: kubeletImageGC,
	}

	aaGroups, err := lookupAntiAffinityGroups(ctx, client, antiAffinityGroups)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.AntiAffinityGroups = aaGroups

	dt, err := lookupDeployTarget(ctx, client, deployTarget)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.DeployTarget = dt

	it, err := lookupInstanceType(ctx, client, instanceType)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.InstanceType = it

	pn, err := lookupPrivateNetworks(ctx, client, privateNetworks)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.PrivateNetworks = pn

	sg, err := lookupSecurityGroups(ctx, client, securityGroups)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.SecurityGroups = sg

	if len(taints) > 0 {
		nodepoolTaints := make(v3.SKSNodepoolTaints)
		for _, t := range taints {
			key, taint, err := parseSKSNodepoolTaint(t)
			if err != nil {
				return nodepoolReq, fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			nodepoolTaints[key] = *taint
		}
		nodepoolReq.Taints = nodepoolTaints
	}

	return nodepoolReq, nil
}

func lookupAntiAffinityGroups(ctx context.Context, client *v3.Client, names []string) ([]v3.AntiAffinityGroup, error) {
	if len(names) == 0 {
		return nil, nil
	}

	groups := make([]v3.AntiAffinityGroup, len(names))
	for i, name := range names {
		antiAffinityGroupList, err := client.ListAntiAffinityGroups(ctx)
		if err != nil {
			return nil, err
		}
		group, err := antiAffinityGroupList.FindAntiAffinityGroup(name)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
		}
		groups[i] = group
	}
	return groups, nil
}

func lookupDeployTarget(ctx context.Context, client *v3.Client, name string) (*v3.DeployTarget, error) {
	if name == "" {
		return nil, nil
	}

	deployTargetList, err := client.ListDeployTargets(ctx)
	if err != nil {
		return nil, err
	}
	deployTarget, err := deployTargetList.FindDeployTarget(name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Deploy Target: %w", err)
	}
	return &deployTarget, nil
}

func lookupInstanceType(ctx context.Context, client *v3.Client, name string) (*v3.InstanceType, error) {
	instanceTypeList, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return nil, err
	}
	instanceType, err := instanceTypeList.FindInstanceTypeByIdOrFamilyAndSize(name)
	if err != nil {
		return nil, fmt.Errorf("error retrieving instance type: %w", err)
	}
	return &instanceType, nil
}

func lookupPrivateNetworks(ctx context.Context, client *v3.Client, names []string) ([]v3.PrivateNetwork, error) {
	if len(names) == 0 {
		return nil, nil
	}

	networks := make([]v3.PrivateNetwork, len(names))
	for i, name := range names {
		networksList, err := client.ListPrivateNetworks(ctx)
		if err != nil {
			return nil, err
		}
		network, err := networksList.FindPrivateNetwork(name)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Private Network: %w", err)
		}
		networks[i] = network
	}
	return networks, nil
}

func lookupSecurityGroups(ctx context.Context, client *v3.Client, names []string) ([]v3.SecurityGroup, error) {
	if len(names) == 0 {
		return nil, nil
	}

	groups := make([]v3.SecurityGroup, len(names))
	for i, name := range names {
		groupsList, err := client.ListSecurityGroups(ctx)
		if err != nil {
			return nil, err
		}
		group, err := groupsList.FindSecurityGroup(name)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Security Group: %w", err)
		}
		groups[i] = group
	}
	return groups, nil
}

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
