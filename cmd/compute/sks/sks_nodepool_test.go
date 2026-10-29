package sks

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"

	"github.com/stretchr/testify/require"
)

func TestParseSKSNodepoolTaint(t *testing.T) {
	testTaints := []struct {
		input         string
		expectedKey   string
		expectedTaint v3.SKSNodepoolTaint
		err           error
	}{
		{
			input:         "key=value:effect",
			expectedKey:   "key",
			expectedTaint: v3.SKSNodepoolTaint{Value: "value", Effect: "effect"},
			err:           nil,
		},
		{
			input:         "exoscale.com/key=value:effect",
			expectedKey:   "exoscale.com/key",
			expectedTaint: v3.SKSNodepoolTaint{Value: "value", Effect: "effect"},
			err:           nil,
		},
		{
			input: "key=:effect",

			err: errExpectedFormatNodepoolTaint,
		},
		{
			input: "key:effect",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: "key=value",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: "=effect",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: "=:effect",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: "=",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: ":",
			err:   errExpectedFormatNodepoolTaint,
		},
		{
			input: "",
			err:   errExpectedFormatNodepoolTaint,
		},
	}

	for _, test := range testTaints {
		k, v, err := parseSKSNodepoolTaint(test.input)
		require.Equal(t, test.err, err)
		if err != nil {
			continue
		}
		require.Equal(t, test.expectedKey, k)
		require.Equal(t, test.expectedTaint, *v)
	}
}

func TestBuildNvidiaMigProfiles(t *testing.T) {
	tests := []struct {
		name      string
		family    v3.InstanceTypeFamily
		profile   string
		expected  *v3.NvidiaMigProfiles
		expectErr bool
	}{
		{
			name:     "empty returns nil regardless of family",
			family:   v3.InstanceTypeFamilyGpua30,
			profile:  "",
			expected: nil,
		},
		{
			name:     "a30 family profile",
			family:   v3.InstanceTypeFamilyGpua30,
			profile:  "4g.24gb",
			expected: &v3.NvidiaMigProfiles{A3024gb: v3.NvidiaMigProfileA3024gb("4g.24gb")},
		},
		{
			name:     "rtxpro6000 family profile",
			family:   v3.InstanceTypeFamilyGpurtx6000pro,
			profile:  "4g.96gb",
			expected: &v3.NvidiaMigProfiles{Rtxpro600096gb: v3.NvidiaMigProfileRtxpro600096gb("4g.96gb")},
		},
		{
			name:     "rtxpro6000 family profile with suffix",
			family:   v3.InstanceTypeFamilyGpurtx6000pro,
			profile:  "2g.48gb+gfx",
			expected: &v3.NvidiaMigProfiles{Rtxpro600096gb: v3.NvidiaMigProfileRtxpro600096gb("2g.48gb+gfx")},
		},
		{
			name:      "profile not valid for family returns error",
			family:    v3.InstanceTypeFamilyGpua30,
			profile:   "4g.96gb",
			expectErr: true,
		},
		{
			name:      "non-GPU family returns error",
			family:    v3.InstanceTypeFamilyStandard,
			profile:   "4g.24gb",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildNvidiaMigProfiles(test.family, test.profile)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, got)
		})
	}
}

func TestBuildCPUManagerConfig(t *testing.T) {
	tests := []struct {
		name      string
		policy    string
		opts      []string
		reconcile string
		kubeRes   *v3.ReservedResources
		sysRes    *v3.ReservedResources
		expected  *v3.CPUManagerConfig
		expectErr bool
	}{
		{
			name:     "no options returns nil",
			expected: nil,
		},
		{
			name:     "empty policy with no other options returns nil",
			policy:   "",
			expected: nil,
		},
		{
			name:      "static policy with options and reservations",
			policy:    "static",
			opts:      []string{"group-by-cores", "dedicated-cpu-id"},
			reconcile: "10s",
			kubeRes:   &v3.ReservedResources{CPU: "200m", Memory: "512Mi"},
			sysRes:    &v3.ReservedResources{CPU: "100m", EphemeralStorage: "1Gi"},
			expected: &v3.CPUManagerConfig{
				CPUManagerPolicy:          v3.CPUManagerConfigCPUManagerPolicyStatic,
				CPUManagerPolicyOptions:   []string{"group-by-cores", "dedicated-cpu-id"},
				CPUManagerReconcilePeriod: "10s",
				KubeReserved:              &v3.ReservedResources{CPU: "200m", Memory: "512Mi"},
				SystemReserved:            &v3.ReservedResources{CPU: "100m", EphemeralStorage: "1Gi"},
			},
		},
		{
			name:     "none policy",
			policy:   "none",
			expected: &v3.CPUManagerConfig{CPUManagerPolicy: v3.CPUManagerConfigCPUManagerPolicyNone},
		},
		{
			name:      "static policy without CPU reservation returns error",
			policy:    "static",
			expectErr: true,
		},
		{
			name:      "static policy with reservation lacking CPU returns error",
			policy:    "static",
			kubeRes:   &v3.ReservedResources{Memory: "512Mi"},
			expectErr: true,
		},
		{
			name:      "reservations without static policy return error",
			kubeRes:   &v3.ReservedResources{CPU: "200m"},
			expectErr: true,
		},
		{
			name:      "reservations with none policy return error",
			policy:    "none",
			sysRes:    &v3.ReservedResources{CPU: "100m"},
			expectErr: true,
		},
		{
			name:      "reconcile period without policy",
			reconcile: "5s",
			expected:  &v3.CPUManagerConfig{CPUManagerReconcilePeriod: "5s"},
		},
		{
			name:      "invalid policy returns error",
			policy:    "bogus",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildCPUManagerConfig(test.policy, test.opts, test.reconcile, test.kubeRes, test.sysRes)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.expected, got)
		})
	}
}

func TestReservedResourcesFromFlags(t *testing.T) {
	t.Run("all empty returns nil", func(t *testing.T) {
		require.Nil(t, reservedResourcesFromFlags("", "", ""))
	})

	t.Run("cpu only", func(t *testing.T) {
		require.Equal(t, &v3.ReservedResources{CPU: "200m"}, reservedResourcesFromFlags("200m", "", ""))
	})

	t.Run("all three", func(t *testing.T) {
		require.Equal(
			t,
			&v3.ReservedResources{CPU: "200m", Memory: "512Mi", EphemeralStorage: "1Gi"},
			reservedResourcesFromFlags("200m", "512Mi", "1Gi"),
		)
	})
}
