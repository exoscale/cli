package sks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksKubeconfigCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"kubeconfig"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	User    string `cli-arg:"#"`

	ExecCredential bool        `cli-short:"x" cli-usage:"output an ExecCredential object to use with a kubeconfig user.exec mode"`
	Groups         []string    `cli-flag:"group" cli-short:"g" cli-usage:"client certificate group. Can be specified multiple times. Defaults to system:masters"`
	Name           string      `cli-short:"n" cli-usage:"cluster name to use in the generated kubeconfig"`
	TTL            int64       `cli-short:"t" cli-usage:"client certificate validity duration in seconds"`
	Zone           v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksKubeconfigCmd) CmdAliases() []string { return []string{"kc"} }

func (c *sksKubeconfigCmd) CmdShort() string {
	return "Generate a Kubernetes kubeconfig file for an SKS cluster"
}

func (c *sksKubeconfigCmd) CmdLong() string {
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

func (c *sksKubeconfigCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksKubeconfigCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.ExecCredential && c.Name != "" {
		return fmt.Errorf("--name cannot be used with --exec-credential")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	// We cannot use the flag's default here as it would be additive
	if len(c.Groups) == 0 {
		c.Groups = []string{"system:masters"}
	}

	req := v3.SKSKubeconfigRequest{
		User:   c.User,
		Groups: c.Groups,
		Ttl:    int64(c.TTL),
	}

	generateResp, err := client.GenerateSKSClusterKubeconfig(ctx, cluster.ID, req)

	if err != nil {
		return fmt.Errorf("error retrieving kubeconfig: %w", err)
	}

	kubeconfig, err := base64.StdEncoding.DecodeString(generateResp.Kubeconfig)
	if err != nil {
		return fmt.Errorf("error decoding kubeconfig content: %w", err)
	}

	if !c.ExecCredential {
		if c.Name != "" {
			kubeconfig, err = setSKSKubeconfigClusterName(kubeconfig, c.Name)
			if err != nil {
				return fmt.Errorf("error rewriting kubeconfig content: %w", err)
			}
		}

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

func setSKSKubeconfigClusterName(kubeconfig []byte, name string) ([]byte, error) {
	if name == "" {
		return kubeconfig, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(kubeconfig, &doc); err != nil {
		return nil, err
	}

	root := yamlDocumentRoot(&doc)
	clusters := yamlMappingValue(root, "clusters")
	if clusters == nil || clusters.Kind != yaml.SequenceNode || len(clusters.Content) == 0 {
		return nil, fmt.Errorf("cluster name not found")
	}

	oldName := ""
	for _, cluster := range clusters.Content {
		clusterName := yamlMappingValue(cluster, "name")
		if clusterName == nil {
			continue
		}

		if oldName == "" {
			oldName = clusterName.Value
		}
		if clusterName.Value == oldName {
			setYAMLDoubleQuotedString(clusterName, name)
		}
	}
	if oldName == "" {
		return nil, fmt.Errorf("cluster name not found")
	}

	contexts := yamlMappingValue(root, "contexts")
	if contexts != nil && contexts.Kind == yaml.SequenceNode {
		for _, context := range contexts.Content {
			contextName := yamlMappingValue(context, "name")
			if contextName != nil && contextName.Value == oldName {
				setYAMLDoubleQuotedString(contextName, name)
			}

			contextValue := yamlMappingValue(context, "context")
			contextCluster := yamlMappingValue(contextValue, "cluster")
			if contextCluster != nil && contextCluster.Value == oldName {
				setYAMLDoubleQuotedString(contextCluster, name)
			}
		}
	}

	currentContext := yamlMappingValue(root, "current-context")
	if currentContext != nil && currentContext.Value == oldName {
		setYAMLDoubleQuotedString(currentContext, name)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func yamlDocumentRoot(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}

	return node
}

func yamlMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

func setYAMLDoubleQuotedString(node *yaml.Node, value string) {
	node.Kind = yaml.ScalarNode
	node.Tag = "!!str"
	node.Value = value
	node.Style = yaml.DoubleQuotedStyle
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksKubeconfigCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
