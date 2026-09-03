package sks

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const sksKubeconfigTestClusterID = "153fcc53-1197-46ae-a8e0-ccf6d09efcb0"

const sksKubeconfigTestContent = `apiVersion: v1
kind: Config
clusters:
- name: 153fcc53-1197-46ae-a8e0-ccf6d09efcb0
  cluster:
    certificate-authority-data: authority
    server: https://153fcc53-1197-46ae-a8e0-ccf6d09efcb0.sks-ch-gva-2.exo.io:443
users:
- name: admin
  user:
    client-certificate-data: certificate
    client-key-data: key
contexts:
- name: 153fcc53-1197-46ae-a8e0-ccf6d09efcb0
  context:
    cluster: 153fcc53-1197-46ae-a8e0-ccf6d09efcb0
    user: admin
current-context: 153fcc53-1197-46ae-a8e0-ccf6d09efcb0
preferences: {}
`

func TestSetSKSKubeconfigClusterNameEmptyNamePreservesContent(t *testing.T) {
	out, err := setSKSKubeconfigClusterName([]byte(sksKubeconfigTestContent), "")
	require.NoError(t, err)
	require.Equal(t, sksKubeconfigTestContent, string(out))
}

func TestSetSKSKubeconfigClusterNameRewritesClusterIdentity(t *testing.T) {
	const name = "my cluster"

	out, err := setSKSKubeconfigClusterName([]byte(sksKubeconfigTestContent), name)
	require.NoError(t, err)

	var kubeconfig struct {
		Clusters []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server string `yaml:"server"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Users []struct {
			Name string `yaml:"name"`
		} `yaml:"users"`
		Contexts []struct {
			Name    string `yaml:"name"`
			Context struct {
				Cluster string `yaml:"cluster"`
				User    string `yaml:"user"`
			} `yaml:"context"`
		} `yaml:"contexts"`
		CurrentContext string `yaml:"current-context"`
	}
	require.NoError(t, yaml.Unmarshal(out, &kubeconfig))

	require.Len(t, kubeconfig.Clusters, 1)
	require.Equal(t, name, kubeconfig.Clusters[0].Name)
	require.Equal(t, "https://"+sksKubeconfigTestClusterID+".sks-ch-gva-2.exo.io:443", kubeconfig.Clusters[0].Cluster.Server)
	require.Len(t, kubeconfig.Users, 1)
	require.Equal(t, "admin", kubeconfig.Users[0].Name)
	require.Len(t, kubeconfig.Contexts, 1)
	require.Equal(t, name, kubeconfig.Contexts[0].Name)
	require.Equal(t, name, kubeconfig.Contexts[0].Context.Cluster)
	require.Equal(t, "admin", kubeconfig.Contexts[0].Context.User)
	require.Equal(t, name, kubeconfig.CurrentContext)

	outString := string(out)
	require.Contains(t, outString, `name: "my cluster"`)
	require.Contains(t, outString, `cluster: "my cluster"`)
	require.Contains(t, outString, `current-context: "my cluster"`)
	require.Contains(t, outString, "server: https://"+sksKubeconfigTestClusterID+".sks-ch-gva-2.exo.io:443")
	require.NotContains(t, outString, "server: https://my cluster")
}

func TestSKSKubeconfigCmdRejectsNameWithExecCredential(t *testing.T) {
	cmd := &sksKubeconfigCmd{
		ExecCredential: true,
		Name:           "my cluster",
	}

	err := cmd.CmdRun(nil, nil)
	require.EqualError(t, err, "--name cannot be used with --exec-credential")
}

func TestSetSKSKubeconfigClusterNameEscapesDoubleQuotedName(t *testing.T) {
	const name = `team "a" cluster`

	out, err := setSKSKubeconfigClusterName([]byte(sksKubeconfigTestContent), name)
	require.NoError(t, err)

	outString := string(out)
	require.Contains(t, outString, `name: "team \"a\" cluster"`)
	require.Contains(t, outString, `current-context: "team \"a\" cluster"`)
}
