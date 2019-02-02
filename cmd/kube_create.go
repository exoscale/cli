package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
	"github.com/cenkalti/backoff"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

const (
	// kubeDockerVersion is the version installed on Ubuntu Xenial
	kubeDockerVersion = "18.06"

	// kubeCalicoVersion is the version of Calico installed
	kubeCalicoVersion = "3.4"

	// kubeDefaultTemplate is the template to install Kubernetes on.
	//
	// Using xenial means cfssl must be installed by other means
	kubeDefaultTemplate = defaultTemplate
)

// kubeCreateDebug represents a debug mode flag
var kubeCreateDebug bool

// kubeBootstrapStep represents a k8s instance bootstrap step
type kubeBootstrapStep struct {
	name    string
	command string
}

type kubeCluster struct {
	Name              string
	KubernetesVersion string
	DockerVersion     string
	CalicoVersion     string
	Address           string
}

// kubeBootstrapSteps represents a k8s instance bootstrap steps
var kubeBootstrapSteps = []kubeBootstrapStep{
	{
		name: "Instance system upgrade",
		command: `\
set -xe

sudo -E DEBIAN_FRONTEND=noninteractive apt-get update
sudo -E DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
sudo -E DEBIAN_FRONTEND=noninteractive apt-get install -y \
	apt-transport-https \
	ca-certificates \
	curl \
	golang-cfssl \
	software-properties-common
nohup sh -c 'sleep 5s ; sudo reboot' &
exit`,
	},
	{
		name: "Docker Engine installation",
		command: `\
set -xe

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

sudo add-apt-repository \
	"deb [arch=amd64] https://download.docker.com/linux/ubuntu \
	$(lsb_release -cs) \
	stable"

sudo -E DEBIAN_FRONTEND=noninteractive apt-get update

PKG_VERSION=$(apt-cache madison docker-ce | awk '$3 ~ /{{ .DockerVersion }}/ { print $3 }' | sort -t : -k 2 -n | tail -n 1)
if [[ -z "${PKG_VERSION}" ]]; then
	echo "error: unable to find docker-ce package for version {{ .DockerVersion }}" >&2
	exit 1
fi

sudo -E DEBIAN_FRONTEND=noninteractive apt-get install -y docker-ce=${PKG_VERSION}
sudo apt-mark hold docker-ce

cat <<EOF > csr.json
{
	"hosts": ["{{ .Address }}"],
	"key": {"algo": "rsa", "size": 2048},
	"names": [{"C": "CH", "L": "Lausanne", "O": "Exoscale", "OU": "exokube", "ST": ""}]
}
EOF

cfssl genkey -initca csr.json | cfssljson -bare ca

cfssl gencert \
	-ca ca.pem \
	-ca-key ca-key.pem \
	-hostname {{ .Address }} csr.json | cfssljson -bare

cat <<EOF | sudo tee /etc/docker/daemon.json
{
	"hosts": ["unix:///var/run/docker.sock", "tcp://0.0.0.0:2376"],
	"tlsverify": true,
	"tlscacert": "/etc/docker/ca.pem",
	"tlscert": "/etc/docker/cert.pem",
	"tlskey": "/etc/docker/key.pem",
	"exec-opts": ["native.cgroupdriver=systemd"],
	"storage-driver": "overlay2",
	"log-driver": "json-file",
	"log-opts": {
		"max-size": "100m"
	}
}
EOF

sudo mv ca.pem /etc/docker/ca.pem
sudo mv cert.pem /etc/docker/cert.pem
sudo mv cert-key.pem /etc/docker/key.pem
rm -f *.{csr,json,pem}

sudo mkdir /etc/systemd/system/docker.service.d/
cat <<EOF | sudo tee /etc/systemd/system/docker.service.d/override.conf
[Service]
ExecStart=
ExecStart=/usr/bin/dockerd
EOF
sudo systemctl daemon-reload \
 && sudo systemctl restart docker`,
	},
	{
		name: "Kubernetes cluster node installation", command: `\
set -xe

curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo -E DEBIAN_FRONTEND=noninteractive apt-get update

PKG_VERSION=$(apt-cache madison kubelet | awk '$3 ~ /{{ .KubernetesVersion }}-/ { print $3 }' | sort -t "-" -k 2 -n | tail -n 1)
if [[ -z "${PKG_VERSION}" ]]; then
	echo "error: unable to find kubelet package for version {{ .KubernetesVersion }}" >&2
	exit 1
fi

sudo -E DEBIAN_FRONTEND=noninteractive apt-get install -y kubelet=${PKG_VERSION} \
	kubeadm=${PKG_VERSION} \
	kubectl=${PKG_VERSION}
sudo apt-mark hold kubelet kubeadm kubectl`,
	}, {
		name: "Kubernetes cluster node initialization",
		command: `\
set -xe

sudo kubeadm init \
	--pod-network-cidr=192.168.0.0/16 \
	--kubernetes-version "{{ .KubernetesVersion }}"
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes --all node-role.kubernetes.io/master-
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf apply \
		-f https://docs.projectcalico.org/v{{ .CalicoVersion }}/getting-started/kubernetes/installation/hosted/etcd.yaml
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf apply \
		-f https://docs.projectcalico.org/v{{ .CalicoVersion }}/getting-started/kubernetes/installation/hosted/calico.yaml`,
	},
}

// kubeCreateCmd represents the create command
var kubeCreateCmd = &cobra.Command{
	Use:   "create <cluster name>",
	Short: "Create and configure a standalone Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		clusterName := args[0]

		vm, err := getVirtualMachineByNameOrID(clusterName)
		if err != nil {
			// If the cluster instance doesn't exist (expected case) we receive a well deserved
			// "431 not found" error, but check if we got any other error just in case...
			if csError, ok := err.(*egoscale.ErrorResponse); ok && csError.ErrorCode != egoscale.ParamError {
				return err
			}
		}
		if vm != nil {
			return fmt.Errorf("cluster instance %q already exists", clusterName)
		}

		kubeCreateDebug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		kubernetesVersion, err := fetchKubernetesVersion()
		if err != nil {
			return err
		}

		sizeOpt, err := cmd.Flags().GetString("size")
		if err != nil {
			return err
		}

		size, err := getServiceOfferingByName(sizeOpt)
		if err != nil {
			return err
		}

		zone, err := getZoneIDByName(gCurrentAccount.DefaultZone)
		if err != nil {
			return err
		}

		sg, err := createExokubeSecurityGroup()
		if err != nil {
			return err
		}

		template, err := getTemplateByName(zone, kubeDefaultTemplate)
		if err != nil {
			return err
		}

		userName, ok := template.Details["username"]
		if !ok {
			return fmt.Errorf("cannot fetch username from template %q", defaultTemplate)
		}

		r, errs := createVM([]egoscale.DeployVirtualMachine{{
			Name:              clusterName,
			ZoneID:            zone,
			ServiceOfferingID: size.ID,
			TemplateID:        template.ID,
			RootDiskSize:      10,
			SecurityGroupIDs:  []egoscale.UUID{*sg.ID},
		}})

		if len(errs) > 0 {
			return errs[0]
		}

		vm = &r[0]

		if _, err := cs.Request(egoscale.CreateTags{
			ResourceType: vm.ResourceType(),
			ResourceIDs:  []egoscale.UUID{*vm.ID},
			Tags: []egoscale.ResourceTag{
				{Key: "managedby", Value: "exokube"},
				{Key: kubeTagName, Value: kubernetesVersion},
			},
		}); err != nil {
			return fmt.Errorf("unable to tag cluster instance: %s", err)
		}

		if err := saveKubeData(clusterName, "instance", []byte(vm.ID.String())); err != nil {
			return fmt.Errorf("unable to write Kubernetes configuration file: %s", err)
		}

		fmt.Println("Bootstrapping Kubernetes cluster (can take up to several minutes):")

		sshClient, err := newSSHClient(
			vm.IP().String(),
			userName,
			getKeyPairPath(vm.ID.String()),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize SSH client: %s", err)
		}

		if err := bootstrapExokubeCluster(sshClient, kubeCluster{
			Name:              clusterName,
			KubernetesVersion: kubernetesVersion,
			CalicoVersion:     kubeCalicoVersion,
			DockerVersion:     kubeDockerVersion,
			Address:           vm.IP().String(),
		}, kubeCreateDebug); err != nil {
			return fmt.Errorf("cluster bootstrap failed: %s", err)
		}

		fmt.Printf(`
Your Kubernetes cluster is ready. What do now?

1. Install the "kubectl" command, if you don't have it already:

    https://kubernetes.io/docs/tasks/tools/install-kubectl/

2. Execute the following command:

    eval $(exo lab kube env "%s")

You might want to persist this change by adding it to your shell startup
configuration (e.g. ~/.bashrc, ~/.zshrc).

3. Check that your cluster is reachable:

    kubectl cluster-info

4. When you're done with your cluster, you can either:
* stop it using the "exo lab kube stop" command
* restart it later using the "exo lab kube start" command
* delete it permanently using the "exo lab kube delete" command
`,
			clusterName)

		return nil
	},
}

// createExokubeSecurityGroup creates the firewall security group to put kube VM instances into, or returns it if it
// already exists.
func createExokubeSecurityGroup() (*egoscale.SecurityGroup, error) {
	sg, err := getSecurityGroupByNameOrID(kubeSecurityGroup)
	if err == nil {
		return sg, err
	}

	r, ok := err.(*egoscale.ErrorResponse)
	if !ok || r.ErrorCode != egoscale.ParamError {
		return nil, err
	}

	resp, err := cs.RequestWithContext(gContext, &egoscale.CreateSecurityGroup{
		Name:        kubeSecurityGroup,
		Description: "Created by exo CLI",
	})

	if err != nil {
		return nil, err
	}

	sg = resp.(*egoscale.SecurityGroup)

	cidrList := []egoscale.CIDR{
		*egoscale.MustParseCIDR("0.0.0.0/0"),
		*egoscale.MustParseCIDR("::/0"),
	}

	sgRules := []egoscale.AuthorizeSecurityGroupIngress{{
		Description: "SSH",
		Protocol:    "TCP",
		StartPort:   22,
		EndPort:     22,
	}, {
		Description: "Docker API",
		Protocol:    "TCP",
		StartPort:   2376,
		EndPort:     2376,
	}, {
		Description: "Kubernetes API",
		Protocol:    "TCP",
		StartPort:   6443,
		EndPort:     6443,
	}}

	for _, rule := range sgRules {
		rule.SecurityGroupID = sg.ID
		rule.CIDRList = cidrList
		if _, err = cs.RequestWithContext(gContext, rule); err != nil {
			return nil, err
		}
	}

	return sg, nil
}

func bootstrapExokubeCluster(sshClient *sshClient, cluster kubeCluster, debug bool) error {
	for _, step := range kubeBootstrapSteps {
		var (
			stdout, stderr io.Writer
			cmd            bytes.Buffer
			errBuf         bytes.Buffer
		)

		stderr = &errBuf
		if debug {
			stdout = os.Stderr
			stderr = os.Stderr
		}

		err := template.Must(template.New("command").Parse(step.command)).Execute(&cmd, cluster)
		if err != nil {
			return fmt.Errorf("template error: %s", err)
		}

		fmt.Printf("%s... ", step.name)

		if err := sshClient.runCommand(cmd.String(), stdout, stderr); err != nil {
			fmt.Printf("\n%s: failed\n", step.name)

			if errBuf.Len() > 0 {
				fmt.Println(errBuf.String())
			}

			return err
		}

		fmt.Printf("success!\n")
	}

	for _, file := range []string{"ca.pem", "cert.pem", "key.pem"} {
		err := sshClient.scp("/etc/docker/"+file, path.Join(getKubeconfigPath(cluster.Name), "docker", file))
		if err != nil {
			return fmt.Errorf("unable to retrieve Docker host file %q: %s", file, err)
		}
	}

	err := sshClient.scp("/etc/kubernetes/admin.conf", path.Join(getKubeconfigPath(cluster.Name), "kubeconfig"))
	if err != nil {
		return fmt.Errorf("unable to retrieve Kubernetes cluster configuration: %s", err)
	}

	return nil
}

type sshClient struct {
	host    string
	hostKey ssh.Signer
	user    string
	c       *ssh.Client
}

func newSSHClient(host, hostUser, keyFile string) (*sshClient, error) {
	var c = sshClient{
		host: host + ":22",
		user: hostUser,
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read cluster instance SSH private key: %s", err)
	}

	if c.hostKey, err = ssh.ParsePrivateKey(key); err != nil {
		return nil, fmt.Errorf("unable to parse cluster instance SSH private key: %s", err)
	}

	return &c, nil
}

func (c *sshClient) runCommand(cmd string, stdout, stderr io.Writer) error {
	var err error

	retryOp := func() error {
		if c.c, err = ssh.Dial("tcp", c.host, &ssh.ClientConfig{
			User:            c.user,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(c.hostKey)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}); err != nil {
			return fmt.Errorf("unable to connect to cluster instance: %s", err)
		}

		sshSession, err := c.c.NewSession()
		if err != nil {
			return fmt.Errorf("unable to create SSH session: %s", err)
		}
		defer sshSession.Close()

		sshSession.Stdout = stdout
		sshSession.Stderr = stderr

		if err := sshSession.Run(cmd); err != nil {
			return err
		}

		return nil
	}

	if err = backoff.RetryNotify(
		retryOp,
		backoff.WithMaxRetries(backoff.NewConstantBackOff(10*time.Second), 6),
		func(_ error, d time.Duration) {
			if kubeCreateDebug {
				fmt.Printf("! Cluster instance not ready yet, retrying in %s...\n", d)
			}
		}); err != nil {
		return err
	}

	return nil
}

func (c *sshClient) scp(src, dst string) error {
	var buf bytes.Buffer

	if err := c.runCommand(fmt.Sprintf("sudo cat %s", src), &buf, nil); err != nil {
		return err
	}

	if _, err := os.Stat(path.Dir(dst)); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(dst), os.ModePerm); err != nil {
			return fmt.Errorf("unable to create directory %q: %s", path.Dir(dst), err)
		}
	}

	return ioutil.WriteFile(dst, buf.Bytes(), 0600)
}

// fetchKubernetesVersion fetches the version from the latest file
//
// https://godoc.org/github.com/kubernetes/kubernetes/cmd/kubeadm/app/util#KubernetesReleaseVersion
func fetchKubernetesVersion() (string, error) {
	r, err := http.Get("https://dl.k8s.io/release/stable.txt")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to find Kubernetes release for stable version")
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	v := strings.TrimSuffix(strings.TrimPrefix(string(b), "v"), "\n")
	_, err = semver.Parse(v)
	return v, err
}

func init() {
	kubeCreateCmd.PersistentFlags().BoolVarP(&kubeCreateDebug, "debug", "d", false, "debug mode on")
	kubeCreateCmd.Flags().StringP("size", "s", "medium", "<name | id> "+
		"(micro|tiny|small|medium|large|extra-large|huge|mega|titan|jumbo)")
	kubeCmd.AddCommand(kubeCreateCmd)
}
