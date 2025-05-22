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
