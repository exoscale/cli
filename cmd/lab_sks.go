package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var labSKSCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
}

func init() {
	labCmd.AddCommand(labSKSCmd)
}

// labSKSFetchTempKubeconfig requests a short-lived (15mn) cluster-admin level kubeconfig from the
// specified SKS cluster, stores it in a temporary file and returns the file path to the caller.
func labSKSFetchTempKubeconfig(zone, c string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "exo-lab-sks-*")
	if err != nil {
		return "", nil
	}

	ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
	cluster, err := lookupSKSCluster(ctx, zone, c)
	if err != nil {
		return "", err
	}

	b64Kubeconfig, err := cluster.RequestKubeconfig(
		"exo-cli",
		[]string{"system:masters"},
		15*time.Minute)
	if err != nil {
		return "", err
	}

	kubeconfig, err := base64.StdEncoding.DecodeString(b64Kubeconfig)
	if err != nil {
		return "", fmt.Errorf("error decoding kubeconfig content: %s", err)
	}

	if _, err = tmpFile.Write(kubeconfig); err != nil {
		return "", fmt.Errorf("error writing kubeconfig content to file: %s", err)
	}

	return tmpFile.Name(), tmpFile.Close()
}
