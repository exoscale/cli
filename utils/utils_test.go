package utils

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/stretchr/testify/assert"
)

func TestParseInstanceType(t *testing.T) {
	testCases := []struct {
		instanceType   string
		expectedFamily v3.InstanceTypeFamily
		expectedSize   v3.InstanceTypeSize
	}{
		{"standard.large", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("large")},
		{"gpu2.mega", v3.InstanceTypeFamily("gpu2"), v3.InstanceTypeSize("mega")},
		{"colossus", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("colossus")},
		{"", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("")},
		{"invalid-format", v3.InstanceTypeFamily("standard"), v3.InstanceTypeSize("invalid-format")},
	}

	for _, tc := range testCases {
		t.Run(tc.instanceType, func(t *testing.T) {
			result := ParseInstanceType(tc.instanceType)
			assert.Equal(t, tc.expectedFamily, result.Family)
			assert.Equal(t, tc.expectedSize, result.Size)
		})
	}
}
