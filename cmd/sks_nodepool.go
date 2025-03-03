package cmd

import (
	"errors"
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

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
