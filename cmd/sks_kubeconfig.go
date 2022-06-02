package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type sksKubeconfigCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"kubeconfig"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	User    string `cli-arg:"#"`

	ExecCredential bool     `cli-short:"x" cli-usage:"output an ExecCredential object to use with a kubeconfig user.exec mode"`
	Groups         []string `cli-flag:"group" cli-short:"g" cli-usage:"client certificate group. Can be specified multiple times. Defaults to system:masters"`
	TTL            int64    `cli-short:"t" cli-usage:"client certificate validity duration in seconds"`
	Zone           string   `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksKubeconfigCmd) cmdAliases() []string { return []string{"kc"} }

func (c *sksKubeconfigCmd) cmdShort() string {
	return "Generate a Kubernetes kubeconfig file for an SKS cluster"
}

func (c *sksKubeconfigCmd) cmdLong() string {
	return `This command generates a kubeconfig file to be used for authenticating to an SKS
cluster API.

The "user" command argument corresponds to the CN field of the generated X.509
client certificate. Optionally, you can specify client certificate groups
using the "-g|--group" option: those groups will be set in the "O" field of
the certificate. See [1] for more information about Kubernetes authentication
certificates.

Example usage:

    # Obtain "cluster-admin" credentials
    $ exo compute sks kubeconfig my-cluster admin \
     	--zone de-fra-1 \
        -g system:masters \
        -t $((86400 * 7)) > $HOME/.kube/my-cluster.config
    $ kubeconfig --kubeconfig=$HOME/.kube/my-cluster.config get pods

Note: if no TTL value is specified, the API applies a default value as a
safety measure. Please look up the API documentation for more information.

## Using exo CLI as Kubernetes credential plugin

If you wish to avoid leaving sensitive credentials on your system, you can use
exo CLI as a Kubernetes client-go credential plugin[2] to generate and return
a kubeconfig dynamically when invoked by kubectl without storing it on disk.

To achieve this configuration, edit your kubeconfig file so that the
"users" section relating to your cluster ("my-sks-cluster" in the following
example) looks like:

    apiVersion: v1
    kind: Config
    clusters:
    - name: my-sks-cluster
      cluster:
        certificate-authority-data: **BASE64-ENCODED CLUSTER CERTIFICATE**
        server: https://153fcc53-1197-46ae-a8e0-ccf6d09efcb0.sks-ch-gva-2.exo.io:443
    users:
    - name: exo@my-sks-cluster
      user:
        # The "exec" section replaces "client-certificate-data"/"client-key-data"
        exec:
          apiVersion: "client.authentication.k8s.io/v1beta1"
          command: exo
          args:
          - sks
          - kubeconfig
          - my-sks-cluster
          - --zone=ch-gva-2
          - --exec-credential
          - user
    contexts:
    - name: my-sks-cluster
      context:
        cluster: my-sks-cluster
        user: exo@my-sks-cluster
    current-context: my-sks-cluster

Notes:

* The "exo" CLI binary must be installed in a directory listed in your PATH
  shell environment variable.
* You can specify the "--group" flag in the user.exec.args section referencing
  a non-admin group to restrict the privileges of the operator using kubectl.

[1]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs
[2]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins
`
}

func (c *sksKubeconfigCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksKubeconfigCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	// We cannot use the flag's default here as it would be additive
	if len(c.Groups) == 0 {
		c.Groups = []string{"system:masters"}
	}

	cluster, err := cs.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		return err
	}

	b64Kubeconfig, err := cs.GetSKSClusterKubeconfig(
		ctx,
		c.Zone,
		cluster,
		c.User,
		c.Groups,
		time.Duration(c.TTL)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("error retrieving kubeconfig: %w", err)
	}

	kubeconfig, err := base64.StdEncoding.DecodeString(b64Kubeconfig)
	if err != nil {
		return fmt.Errorf("error decoding kubeconfig content: %w", err)
	}

	if !c.ExecCredential {
		fmt.Print(string(kubeconfig))
		return nil
	}

	k := struct {
		Users []struct {
			Name string            `yaml:"name"`
			User map[string]string `yaml:"user"`
		} `yaml:"users"`
	}{}
	if err := yaml.Unmarshal(kubeconfig, &k); err != nil {
		return fmt.Errorf("error decoding kubeconfig content: %w", err)
	}

	ecClientCertificateData, err := base64.StdEncoding.DecodeString(k.Users[0].User["client-certificate-data"])
	if err != nil {
		return fmt.Errorf("error decoding kubeconfig content: %w", err)
	}

	ecClientKeyData, err := base64.StdEncoding.DecodeString(k.Users[0].User["client-key-data"])
	if err != nil {
		return fmt.Errorf("error decoding kubeconfig content: %w", err)
	}

	ecOut, err := json.Marshal(map[string]interface{}{
		"apiVersion": "client.authentication.k8s.io/v1beta1",
		"kind":       "ExecCredential",
		"status": map[string]string{
			"clientCertificateData": string(ecClientCertificateData),
			"clientKeyData":         string(ecClientKeyData),
		},
	})
	if err != nil {
		return fmt.Errorf("error encoding exec credential content: %w", err)
	}

	fmt.Print(string(ecOut))
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksKubeconfigCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksKubeconfigCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
