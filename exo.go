package main

import (
	"egoscale"
	"os"
	"fmt"
)

func main() {

	endpoint := os.Getenv("EXOSCALE_ENDPOINT")
	apiKey := os.Getenv("EXOSCALE_API_KEY")
	apiSecret:= os.Getenv("EXOSCALE_API_SECRET")
	client := egoscale.NewClient(endpoint, apiKey, apiSecret)

	topo, err := client.GetTopology()
	if err != nil {
		fmt.Printf("got error: %+v\n", err)
		return
	}

	rules := []egoscale.SecurityGroupRule{
		{
			SecurityGroupId: "",
			Cidr: "0.0.0.0/0",
			Protocol: "TCP",
			Port: 22,
		},
		{
			SecurityGroupId: "",
			Cidr: "0.0.0.0/0",
			Protocol: "TCP",
			Port: 2376,
		},
		{
			SecurityGroupId: "",
			Cidr: "0.0.0.0/0",
			Protocol: "ICMP",
			IcmpType: 8,
			IcmpCode: 0,
		},
	}

	sgid, present := topo.SecurityGroups["egoscale"]
	if !present {
		resp, err := client.CreateSecurityGroupWithRules("egoscale", rules, make([]egoscale.SecurityGroupRule,0,0))
		if err != nil {
			fmt.Printf("got error: %+v\n", err)
			return
		}
		sgid = resp.Id
	}

	tags := make(map[string]string)
	tags["docker-machine"] = "true"
	profile := egoscale.MachineProfile{
		Template: topo.Images["ubuntu-14.04"][10],
		ServiceOffering: topo.Profiles["large"],
		SecurityGroups: []string{ sgid,},
		Keypair: topo.Keypairs[0],
		Userdata: "#cloud-config\nmanage_etc_hosts: true\nfqdn: deployed-by-egoscale\n",
		Zone: topo.Zones["ch-gva-2"],
		Name: "deployed-by-egoscale",
		Tags: tags,
	}

	resp, err := client.CreateVirtualMachine(profile)

	if err != nil {
		fmt.Printf("got error: %+v\n", err)
		return
	}

	fmt.Printf("got reply: %s\n", resp)
}
