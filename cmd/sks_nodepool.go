package cmd

import (
	"errors"
	"regexp"

	egoscale "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var sksNodepoolCmd = &cobra.Command{
	Use:     "nodepool",
	Short:   "Manage SKS cluster Nodepools",
	Aliases: []string{"np"},
}

// parseSKSNodepoolTaint parses a CLI-formatted Kubernetes Node taint
// description formatted as KEY=VALUE:EFFECT, and returns descrete values
// for the taint key as well as the value/effect as egoscale.SKSNodepoolTaint,
// or an error if the input value parsing failed.
func parseSKSNodepoolTaint(v string) (string, *egoscale.SKSNodepoolTaint, error) {
	res := regexp.MustCompile(`(\w+)=(\w+):(\w+)`).FindStringSubmatch(v)
	if len(res) != 4 {
		return "", nil, errors.New("expected format KEY=VALUE:EFFECT")
	}
	taintKey, taintValue, taintEffect := res[1], res[2], res[3]

	if taintKey == "" || taintValue == "" || taintEffect == "" {
		return "", nil, errors.New("expected format KEY=VALUE:EFFECT")
	}

	return taintKey, &egoscale.SKSNodepoolTaint{Effect: taintEffect, Value: taintValue}, nil
}

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
