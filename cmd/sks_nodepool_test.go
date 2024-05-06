package cmd

import (
	"testing"

	egoscale "github.com/exoscale/egoscale/v2"
	"github.com/stretchr/testify/require"
)

func TestParseSKSNodepoolTaint(t *testing.T) {
	testTaints := []struct {
		input         string
		expectedKey   string
		expectedTaint egoscale.SKSNodepoolTaint
		err           error
	}{
		{
			input:         "key=value:effect",
			expectedKey:   "key",
			expectedTaint: egoscale.SKSNodepoolTaint{Value: "value", Effect: "effect"},
			err:           nil,
		},
		{
			input:         "key=:effect",
			expectedKey:   "key",
			expectedTaint: egoscale.SKSNodepoolTaint{Effect: "effect"},
			err:           nil,
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
