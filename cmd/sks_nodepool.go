package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	egoscale "github.com/exoscale/egoscale/v2"
)

var sksNodepoolCmd = &cobra.Command{
	Use:     "nodepool",
	Short:   "Manage SKS cluster Nodepools",
	Aliases: []string{"np"},
}
var errExpectedFormatNodepoolTaint = errors.New("expected format KEY=VALUE:EFFECT or KEY=:EFFECT")

// parseSKSNodepoolTaint parses a CLI-formatted Kubernetes Node taint.
// According to:
// https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#taint
// We will support only: KEY=VALUE:EFFECT or KEY=:EFFECT for the moment.
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

	if taintKey == "" || taintEffect == "" {
		return "", nil, errExpectedFormatNodepoolTaint
	}

	return taintKey, &egoscale.SKSNodepoolTaint{Effect: taintEffect, Value: taintValue}, nil
}

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
