package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/exoscale/egoscale"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// kubeCreateDebug represents a debug mode flag
var kubeCreateDebug bool

// kubeBootstrapStep represents a k8s instance bootstrap step
type kubeBootstrapStep struct {
	name    string
	command string
}

type kubeCluster struct {
	Name    string
	Version string
	Address string
}

// kubeBootstrapSteps represents a k8s instance bootstrap steps
var kubeBootstrapSteps = []kubeBootstrapStep{
	{name: "Instance system upgrade", command: `\
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
	curl \
	golang-cfssl \
	software-properties-common
nohup sh -c 'sleep 5s ; sudo reboot' &
exit`},
	{name: "Docker Engine installation", command: `\
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
sudo apt-get update && sudo apt-get install -y docker-ce=18.06.0~ce~3-0~ubuntu

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

sudo sed -i -re 's#^ExecStart=.*#ExecStart=/usr/bin/dockerd#' /lib/systemd/system/docker.service
sudo systemctl daemon-reload && sudo systemctl restart docker`},
	{name: "Kubernetes cluster node installation", command: `\
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

sudo apt-get update && \
sudo apt-get install -y kubelet kubeadm kubectl && \
sudo apt-mark hold kubelet kubeadm kubectl`},
	{name: "Kubernetes cluster node initialization", command: `\
sudo kubeadm init --pod-network-cidr=192.168.0.0/16 --kubernetes-version {{ .Version }} &&
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes --all node-role.kubernetes.io/master- &&
sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf apply \
  -f https://docs.projectcalico.org/v3.3/getting-started/kubernetes/installation/hosted/rbac-kdd.yaml \
  -f https://docs.projectcalico.org/v3.3/getting-started/kubernetes/installation/hosted/kubernetes-datastore/calico-networking/1.7/calico.yaml`},
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

		kubeCreateDebug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		version, err := cmd.Flags().GetString("version")
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

		template, _ := getTemplateByName(zone, defaultTemplate)

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

		vm := r[0]

		if _, err := cs.Request(egoscale.CreateTags{
			ResourceType: vm.ResourceType(),
			ResourceIDs:  []egoscale.UUID{*vm.ID},
			Tags:         []egoscale.ResourceTag{{Key: kubeInstanceTagKey, Value: kubeInstanceTagValue}},
		}); err != nil {
			return fmt.Errorf("unable to tag cluster instance: %s", err)
		}

		fmt.Println("🚧 Bootstrapping Kubernetes cluster (can take up to several minutes):")

		sshClient, err := newSSHClient(
			vm.IP().String(),
			"ubuntu",
			path.Join(getKeyPairPath(vm.ID.String()), "id_rsa"),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize SSH client: %s", err)
		}

		if err := bootstrapExokubeCluster(sshClient, kubeCluster{
			Name:    clusterName,
			Version: version,
			Address: vm.IP().String(),
		}, kubeCreateDebug); err != nil {
			return fmt.Errorf("Cluster bootstrap failed: %s", err) // nolint: golint
		}

		if err := saveKubeData(clusterName, "instance", []byte(vm.ID.String())); err != nil {
			return fmt.Errorf("unable to write Kubernetes configuration file: %s", err)
		}

		fmt.Printf(`
🏁 Your Kubernetes cluster is ready. What to do now?

1. Install the "kubectl" command, if you don't have it already:

    https://kubernetes.io/docs/tasks/tools/install-kubectl/

2. Execute the following command:

    eval $(exo kube env "%s")

You might want to persist this change by adding it to your shell startup
configuration (e.g. ~/.bashrc, ~/.zshrc).

3. Check that your cluster is reachable:

    kubectl cluster-info

4. Profit! ✨🦄🌈

5. When you're done with your cluster, you can either stop it using the
"exo kube stop" command and restart it later using the "exo kube start"
command, or delete it permanently using the "exo kube delete" command.
`,
			clusterName)

		return nil
	},
}

// createExokubeSecurityGroup creates the firewall security group to put kube VM instances into, or returns it if it
// already exists.
func createExokubeSecurityGroup() (*egoscale.SecurityGroup, error) {
	var (
		sg  *egoscale.SecurityGroup
		err error
	)

	if sg, err = getSecurityGroupByNameOrID(kubeSecurityGroup); err != nil {
		if r, ok := err.(*egoscale.ErrorResponse); ok {
			// Looks like the SG doesn't exist, try to create it
			if r.ErrorCode == egoscale.ParamError {
				resp, err := cs.RequestWithContext(gContext, &egoscale.CreateSecurityGroup{
					Name:        kubeSecurityGroup,
					Description: "Created by exo CLI",
				})
				if err != nil {
					return nil, err
				}

				sg = resp.(*egoscale.SecurityGroup)

				sgRules := []egoscale.AuthorizeSecurityGroupIngress{
					{
						SecurityGroupID: sg.ID,
						Description:     "SSH",
						CIDRList:        []egoscale.CIDR{*egoscale.MustParseCIDR("0.0.0.0/0")},
						Protocol:        "TCP",
						StartPort:       22,
						EndPort:         22,
					},
					{
						SecurityGroupID: sg.ID,
						Description:     "Docker API",
						CIDRList:        []egoscale.CIDR{*egoscale.MustParseCIDR("0.0.0.0/0")},
						Protocol:        "TCP",
						StartPort:       2376,
						EndPort:         2376,
					},
					{
						SecurityGroupID: sg.ID,
						Description:     "Kubernetes API",
						CIDRList:        []egoscale.CIDR{*egoscale.MustParseCIDR("0.0.0.0/0")},
						Protocol:        "TCP",
						StartPort:       6443,
						EndPort:         6443,
					},
				}

				for _, rule := range sgRules {
					if _, err = cs.RequestWithContext(gContext, rule); err != nil {
						return nil, err
					}
				}

				return sg, nil
			}
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
			w              *wow.Wow
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

		if !kubeCreateDebug {
			w = wow.New(os.Stdout, spin.Get(spin.Dots), " "+step.name)
			w.Start()
		} else {
			fmt.Println(">>>", step.name)
		}

		if err := sshClient.runCommand(cmd.String(), stdout, stderr); err != nil {
			if !kubeCreateDebug {
				w.PersistWith(spin.Spinner{Frames: []string{"⚠️"}}, fmt.Sprintf(" %s: failed", step.name))
			}

			if errBuf.Len() > 0 {
				fmt.Println(errBuf.String())
			}

			return err
		}

		if !kubeCreateDebug {
			w.PersistWith(spin.Spinner{Frames: []string{"✅"}}, " "+step.name)
		}
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

func init() {
	kubeCreateCmd.PersistentFlags().BoolVarP(&kubeCreateDebug, "debug", "d", false, "debug mode on")
	kubeCreateCmd.Flags().StringP("version", "v", "stable-1", "<version label> "+
		"(see https://godoc.org/github.com/kubernetes/kubernetes/cmd/kubeadm/app/util#KubernetesReleaseVersion)")
	kubeCreateCmd.Flags().StringP("size", "s", "small", "<name | id> "+
		"(micro|tiny|small|medium|large|extra-large|huge|mega|titan|jumbo)")
	kubeCmd.AddCommand(kubeCreateCmd)
}
