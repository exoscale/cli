package sks

import (
	"context"
	"errors"
	"fmt"
	"slices"
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

type CreateNodepoolOpts struct {
	Name               string
	Description        string
	DiskSize           int64
	InstancePrefix     string
	Size               int64
	InstanceType       string
	Labels             map[string]string
	AntiAffinityGroups []string
	DeployTarget       string
	PrivateNetworks    []string
	SecurityGroups     []string
	Taints             []string
	KubeletImageGC     *v3.KubeletImageGC
	PublicIPAssignment v3.CreateSKSNodepoolRequestPublicIPAssignment

	// Nvidia MIG (Multi-Instance GPU) profile to enable on the Nodepool GPUs.
	// The GPU family it applies to is inferred from the Nodepool instance type.
	// Empty means "leave unset".
	NvidiaMigProfile string
}

// nvidiaMigProfilesA30 and nvidiaMigProfilesRtxpro6000 list the MIG profiles
// supported by each GPU instance type family.
var (
	nvidiaMigProfilesA30 = []v3.NvidiaMigProfileA3024gb{
		v3.NvidiaMigProfileA3024gb1G6Gb,
		v3.NvidiaMigProfileA3024gb1G6GbMe,
		v3.NvidiaMigProfileA3024gb2G12Gb,
		v3.NvidiaMigProfileA3024gb2G12GbMe,
		v3.NvidiaMigProfileA3024gb4G24Gb,
	}

	nvidiaMigProfilesRtxpro6000 = []v3.NvidiaMigProfileRtxpro600096gb{
		v3.NvidiaMigProfileRtxpro600096gb1G24Gb,
		v3.NvidiaMigProfileRtxpro600096gb1G24GbMe,
		v3.NvidiaMigProfileRtxpro600096gb1G24GbGfx,
		v3.NvidiaMigProfileRtxpro600096gb1G24GbMeAll,
		v3.NvidiaMigProfileRtxpro600096gb1G24GbNoMe,
		v3.NvidiaMigProfileRtxpro600096gb2G48Gb,
		v3.NvidiaMigProfileRtxpro600096gb2G48GbGfx,
		v3.NvidiaMigProfileRtxpro600096gb2G48GbMeAll,
		v3.NvidiaMigProfileRtxpro600096gb2G48GbNoMe,
		v3.NvidiaMigProfileRtxpro600096gb4G96Gb,
		v3.NvidiaMigProfileRtxpro600096gb4G96GbGfx,
	}
)

// migProfileValues renders a slice of typed MIG profile enums as plain strings,
// for use in error messages.
func migProfileValues[T ~string](profiles []T) []string {
	values := make([]string, len(profiles))
	for i, p := range profiles {
		values[i] = string(p)
	}
	return values
}

// buildNvidiaMigProfiles resolves the requested MIG (Multi-Instance GPU) profile
// to the egoscale profiles object, selecting the GPU family from the Nodepool
// instance type family. It returns nil when profile is empty, and an error when
// the instance type family has no MIG support or the value is not valid for it.
func buildNvidiaMigProfiles(family v3.InstanceTypeFamily, profile string) (*v3.NvidiaMigProfiles, error) {
	if profile == "" {
		return nil, nil
	}

	switch family {
	case v3.InstanceTypeFamilyGpua30:
		p := v3.NvidiaMigProfileA3024gb(profile)
		if !slices.Contains(nvidiaMigProfilesA30, p) {
			return nil, fmt.Errorf(
				"invalid Nvidia MIG profile %q for instance type family %q, supported values: %s",
				profile, family, strings.Join(migProfileValues(nvidiaMigProfilesA30), ", "),
			)
		}
		return &v3.NvidiaMigProfiles{A3024gb: p}, nil

	case v3.InstanceTypeFamilyGpurtx6000pro:
		p := v3.NvidiaMigProfileRtxpro600096gb(profile)
		if !slices.Contains(nvidiaMigProfilesRtxpro6000, p) {
			return nil, fmt.Errorf(
				"invalid Nvidia MIG profile %q for instance type family %q, supported values: %s",
				profile, family, strings.Join(migProfileValues(nvidiaMigProfilesRtxpro6000), ", "),
			)
		}
		return &v3.NvidiaMigProfiles{Rtxpro600096gb: p}, nil

	default:
		return nil, fmt.Errorf("instance type family %q does not support Nvidia MIG profiles", family)
	}
}

func createNodepoolRequest(
	ctx context.Context,
	client *v3.Client,
	opts CreateNodepoolOpts,
) (v3.CreateSKSNodepoolRequest, error) {

	nodepoolReq := v3.CreateSKSNodepoolRequest{
		Description:        opts.Description,
		DiskSize:           opts.DiskSize,
		InstancePrefix:     opts.InstancePrefix,
		Name:               opts.Name,
		Size:               opts.Size,
		Labels:             opts.Labels,
		KubeletImageGC:     opts.KubeletImageGC,
		PublicIPAssignment: opts.PublicIPAssignment,
	}

	aaGroups, err := lookupAntiAffinityGroups(ctx, client, opts.AntiAffinityGroups)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.AntiAffinityGroups = aaGroups

	dt, err := lookupDeployTarget(ctx, client, opts.DeployTarget)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.DeployTarget = dt

	it, err := lookupInstanceType(ctx, client, opts.InstanceType)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.InstanceType = it

	pn, err := lookupPrivateNetworks(ctx, client, opts.PrivateNetworks)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.PrivateNetworks = pn

	sg, err := lookupSecurityGroups(ctx, client, opts.SecurityGroups)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.SecurityGroups = sg

	if len(opts.Taints) > 0 {
		nodepoolTaints := make(v3.SKSNodepoolTaints)
		for _, t := range opts.Taints {
			key, taint, err := parseSKSNodepoolTaint(t)
			if err != nil {
				return nodepoolReq, fmt.Errorf("invalid taint value %q: %w", t, err)
			}
			nodepoolTaints[key] = *taint
		}
		nodepoolReq.Taints = nodepoolTaints
	}

	migProfiles, err := buildNvidiaMigProfiles(it.Family, opts.NvidiaMigProfile)
	if err != nil {
		return nodepoolReq, err
	}
	nodepoolReq.NvidiaMigProfiles = migProfiles

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
