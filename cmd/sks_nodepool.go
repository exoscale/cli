package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	egoscale "github.com/exoscale/egoscale/v2"
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
func parseSKSNodepoolTaint(v string) (string, *egoscale.SKSNodepoolTaint, error) {
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

	return taintKey, &egoscale.SKSNodepoolTaint{Effect: taintEffect, Value: taintValue}, nil
}

func parseSKSNodepoolTaintV3(v string) (string, *v3.SKSNodepoolTaint, error) {
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

	if l := len(antiAffinityGroups); l > 0 {
		nodepoolReq.AntiAffinityGroups = make([]v3.AntiAffinityGroup, l)
		for i, v := range antiAffinityGroups {
			antiAffinityGroupList, err := client.ListAntiAffinityGroups(ctx)
			if err != nil {
				return nodepoolReq, err
			}
			aaG, err := antiAffinityGroupList.FindAntiAffinityGroup(v)
			if err != nil {
				return nodepoolReq, fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			nodepoolReq.AntiAffinityGroups[i] = aaG
		}
	}

	if deployTarget != "" {
		deployTargetList, err := client.ListDeployTargets(ctx)
		if err != nil {
			return nodepoolReq, err
		}
		deployTarget, err := deployTargetList.FindDeployTarget(deployTarget)
		if err != nil {
			return nodepoolReq, fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		nodepoolReq.DeployTarget = &deployTarget
	}

	nodepoolInstanceTypeList, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolInstanceType, err := nodepoolInstanceTypeList.FindInstanceTypeByIdOrFamilyAndSize(instanceType)
	if err != nil {
		return nodepoolReq, fmt.Errorf("error retrieving instance type: %w", err)
	}
	nodepoolReq.InstanceType = &nodepoolInstanceType

	if l := len(privateNetworks); l > 0 {
		nodepoolPrivateNetworks := make([]v3.PrivateNetwork, l)
		for i, v := range privateNetworks {
			privateNetworksList, err := client.ListPrivateNetworks(ctx)
			if err != nil {
				return nodepoolReq, err
			}
			privateNetwork, err := privateNetworksList.FindPrivateNetwork(v)
			if err != nil {
				return nodepoolReq, fmt.Errorf("error retrieving Private Network: %w", err)
			}
			nodepoolPrivateNetworks[i] = privateNetwork
		}
		nodepoolReq.PrivateNetworks = nodepoolPrivateNetworks
	}

	if l := len(securityGroups); l > 0 {
		nodepoolSecurityGroups := make([]v3.SecurityGroup, l)
		for i, v := range securityGroups {
			securityGroupList, err := client.ListSecurityGroups(ctx)
			if err != nil {
				return nodepoolReq, err
			}
			securityGroup, err := securityGroupList.FindSecurityGroup(v)
			if err != nil {
				return nodepoolReq, fmt.Errorf("error retrieving Security Group: %w", err)
			}
			nodepoolSecurityGroups[i] = securityGroup
		}
		nodepoolReq.SecurityGroups = nodepoolSecurityGroups
	}

	if len(taints) > 0 {
		nodepoolTaints := make(v3.SKSNodepoolTaints)
		for _, t := range taints {
			key, taint, err := parseSKSNodepoolTaintV3(t)
			if err != nil {
				return nodepoolReq, fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			nodepoolTaints[key] = *taint
		}
		nodepoolReq.Taints = nodepoolTaints
	}

	return nodepoolReq, nil
}

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
