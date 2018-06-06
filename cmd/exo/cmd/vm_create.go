package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/exoscale/egoscale/cmd/exo/utils"
	"github.com/spf13/cobra"
)

var templateName = "Linux Debian 9"

// vmCreateCmd represents the create command
var vmCreateCmd = &cobra.Command{
	Use:   "create <vm name>",
	Short: "Create and deploy a virtual machine",
}

func vmCreateRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		vmCreateCmd.Usage()
		return
	}

	userDataPath, err := cmd.Flags().GetString("cloud-init-file")
	if err != nil {
		log.Fatal(err)
	}

	userData := ""

	if userDataPath != "" {
		userData, err = getUserData(userDataPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		log.Fatal(err)
	}

	zone, err = getZoneIDByName(cs, zone)
	if err != nil {
		log.Fatal(err)
	}

	template, err := cmd.Flags().GetString("template")
	if err != nil {
		log.Fatal(err)
	}

	diskSize, err := cmd.Flags().GetInt64("disk")
	if err != nil {
		log.Fatal(err)
	}

	template, err = getTemplateIDByName(cs, template, zone)
	if err != nil {
		log.Fatal(err)
	}

	keypair, err := cmd.Flags().GetString("keypair")
	if err != nil {
		log.Fatal(err)
	}

	sg, err := cmd.Flags().GetString("security-group")
	if err != nil {
		log.Fatal(err)
	}

	sgs, err := getSecuGrpList(cs, sg)
	if err != nil {
		log.Fatal(err)
	}

	ipv6, err := cmd.Flags().GetBool("ipv6")
	if err != nil {
		log.Fatal(err)
	}

	privnet, err := cmd.Flags().GetString("privnet")
	if err != nil {
		log.Fatal(err)
	}

	pvs, err := getPrivnetList(cs, privnet, zone)
	if err != nil {
		log.Fatal(err)
	}

	servOffering, err := cmd.Flags().GetString("service-offering")
	if err != nil {
		log.Fatal(err)
	}

	servOffering, err = getServiceOfferingIDByName(cs, servOffering)
	if err != nil {
		log.Fatal(err)
	}

	affinitygroup, err := cmd.Flags().GetString("anti-affinity-group")
	if err != nil {
		log.Fatal(err)
	}

	affinitygroups, err := getAffinityGroup(cs, affinitygroup)
	if err != nil {
		log.Fatal(err)
	}

	vmInfo := &egoscale.DeployVirtualMachine{
		Name:              args[0],
		UserData:          userData,
		ZoneID:            zone,
		TemplateID:        template,
		RootDiskSize:      diskSize,
		KeyPair:           keypair,
		SecurityGroupIDs:  sgs,
		IP6:               &ipv6,
		NetworkIDs:        pvs,
		ServiceOfferingID: servOffering,
		AffinityGroupIDs:  affinitygroups,
	}

	r, err := createVM(vmInfo)
	if err != nil {
		log.Fatal(err)
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "IP", "ID"})

	table.Append([]string{r.Name, r.IP().String(), r.ID})
	table.Render()
}

func getCommaflag(p string) []string {
	if p == "" {
		return nil
	}

	p = strings.Trim(p, ",")
	args := strings.Split(p, ",")

	res := []string{}

	for _, arg := range args {
		if arg == "" {
			continue
		}
		res = append(res, strings.TrimSpace(arg))
	}

	return res
}

func getSecuGrpList(cs *egoscale.Client, commaParameter string) ([]string, error) {

	sgs := getCommaflag(commaParameter)

	for i, sg := range sgs {
		s, err := getSecuGrpWithNameOrID(cs, sg)
		if err != nil {
			return nil, err
		}
		sgs[i] = s.ID
	}

	return sgs, nil
}

func getPrivnetList(cs *egoscale.Client, commaParameter, zoneID string) ([]string, error) {

	sgs := getCommaflag(commaParameter)

	for i, sg := range sgs {
		s, err := getNetworkIDByName(cs, sg, zoneID)
		if err != nil {
			return nil, err
		}
		sgs[i] = s
	}

	return sgs, nil
}

func getAffinityGroup(cs *egoscale.Client, commaParameter string) ([]string, error) {
	affs := getCommaflag(commaParameter)

	for i, aff := range affs {
		s, err := getAffinityGroupIDByName(cs, aff)
		if err != nil {
			return nil, err
		}
		affs[i] = s
	}

	return affs, nil
}

func getUserData(userDataPath string) (string, error) {
	buff, err := ioutil.ReadFile(userDataPath)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buff), nil
}

func createVM(vmInfos *egoscale.DeployVirtualMachine) (*egoscale.VirtualMachine, error) {
	isDefaultKeyPair := false
	var keyPairs *egoscale.SSHKeyPair

	if vmInfos.KeyPair == "" {
		isDefaultKeyPair = true
		println("Creating sshkey")
		sshKeyName, err := utils.RandStringBytes(64)
		if err != nil {
			return nil, err
		}
		keyPairs, err = createSSHKey(sshKeyName)
		if err != nil {
			r := err.(*egoscale.ErrorResponse)
			if r.ErrorCode != egoscale.ParamError && r.CSErrorCode != egoscale.InvalidParameterValueException {
				return nil, err
			}
			return nil, fmt.Errorf("An sshkey with name %q already exist, please create your VM with another name", sshKeyName)
		}

		defer deleteSSHKey(keyPairs.Name)

		vmInfos.KeyPair = keyPairs.Name

	}

	virtualMachine := &egoscale.VirtualMachine{}
	var errorReq error
	print("Deploying")
	cs.AsyncRequest(vmInfos, func(jobResult *egoscale.AsyncJobResult, err error) bool {

		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus == egoscale.Success {

			if err := jobResult.Response(virtualMachine); err != nil {
				errorReq = err
			}

			println("")
			return false
		}
		fmt.Printf(".")
		return true
	})

	if errorReq != nil {
		return nil, errorReq
	}

	if isDefaultKeyPair {
		saveKeyPair(keyPairs, virtualMachine.ID)
	}

	return virtualMachine, nil
}

func saveKeyPair(keyPairs *egoscale.SSHKeyPair, vmID string) {
	filePath := path.Join(configFolder, "instances", vmID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	filePath = path.Join(filePath, "id_rsa")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(keyPairs.PrivateKey), 0600); err != nil {
			log.Fatalf("SSH private key could not be written. %s", err)
		}
	}
}

func init() {
	vmCreateCmd.Run = vmCreateRun
	vmCreateCmd.Flags().StringP("cloud-init-file", "f", "", "Deploy instance with a cloud-init file")
	vmCreateCmd.Flags().StringP("zone", "z", "ch-dk-2", "<zone name | id | keyword> (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1)")
	vmCreateCmd.Flags().StringP("template", "t", "Linux Ubuntu 18.04", "<template name | id>")
	vmCreateCmd.Flags().Int64P("disk", "d", 50, "<disk size>")
	vmCreateCmd.Flags().StringP("keypair", "k", "", "<ssh keys name>")
	vmCreateCmd.Flags().StringP("security-group", "s", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().StringP("privnet", "p", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().StringP("anti-affinity-group", "a", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().BoolP("ipv6", "6", false, "enable ipv6")
	vmCreateCmd.Flags().StringP("service-offering", "o", "Small", "<name | id> (micro|tiny|small|medium|large|extra-large|huge|mega|titan")
	vmCmd.AddCommand(vmCreateCmd)
}
