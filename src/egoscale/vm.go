package egoscale

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

func (exo *Client) CreateVirtualMachine(p MachineProfile) (string, error) {

	params := url.Values{}
	params.Set("serviceofferingid", p.ServiceOffering)
	params.Set("templateid", p.Template)
	params.Set("zoneid", p.Zone)

	params.Set("displayname", p.Name)
	if len(p.Userdata) > 0 {
		params.Set("userdata", base64.StdEncoding.EncodeToString([]byte(p.Userdata)))
	}
	if len(p.Keypair) > 0 {
		params.Set("keypair", p.Keypair)
	}

	params.Set("securitygroupids", strings.Join(p.SecurityGroups, ","))

	resp, err := exo.Request("deployVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r DeployVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}
