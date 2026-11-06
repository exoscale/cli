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
