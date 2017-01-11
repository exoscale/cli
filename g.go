package main

import (
	"egoscale"
	"fmt"
	"os"
)

func main() {

	endpoint := os.Getenv("EXOSCALE_ENDPOINT")
	apiKey := os.Getenv("EXOSCALE_API_KEY")
	apiSecret := os.Getenv("EXOSCALE_API_SECRET")
	client := egoscale.NewClient(endpoint, apiKey, apiSecret)

	topo, err := client.GetTopology()
	if err != nil {
		fmt.Printf("got error: %+v\n", err)
		return
	}

        list := []egoscale.UserSecurityGroup{
                            {
                                Account: "antoine.coetsier@gmail.com",
                                Group: "default",
                            },
                            {
                                Account: "antoine.coetsier@gmail.com",
                                Group: "web",
                            },
                    }

	rules := []egoscale.SecurityGroupRule{
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            22,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            2376,
		},
                {
			SecurityGroupId: "d61baf10-4cd1-4b69-9e61-424c8008334f",
                        UserSecurityGroupList: list,
			Protocol:        "TCP",
			Port:            443,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "ICMP",
			IcmpType:        8,
			IcmpCode:        0,
		},
	}

	sgid, present := topo.SecurityGroups["egoscale5"]
	if !present {
		resp, err := client.CreateSecurityGroupWithRules("egoscale5", rules, make([]egoscale.SecurityGroupRule, 0, 0))
		if err != nil {
			fmt.Printf("got error: %+v\n", err)
			return
		}
		sgid = resp.Id
	}
        
        fmt.Printf("Security Group ID :%v\n", sgid)
}
