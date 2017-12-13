package egoscale

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ListNics lists the NIC of a VM
func (exo *Client) ListNics(virtualMachineId string) ([]*Nic, error) {
	params := url.Values{}
	params.Set("virtualmachineid", virtualMachineId)

	resp, err := exo.Request("listNics", params)
	if err != nil {
		return nil, err
	}

	var r ListNicsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.Nic, nil
}

// AddIpToNic adds the IP address to the given NIC
func (exo *Client) AddIpToNic(nicId string, ipAddress string, async AsyncInfo) (*NicSecondaryIp, error) {
	params := url.Values{}
	params.Set("nicid", nicId)
	params.Set("ipaddress", ipAddress)

	resp, err := exo.AsyncRequest("addIpToNic", params, async)
	if err != nil {
		return nil, err
	}

	var r AddIpToNicResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.NicSecondaryIp, nil
}

// RemoveIpFromNic removes the IP address (by Id) from the NIC
func (exo *Client) RemoveIpFromNic(ipAddressId string, async AsyncInfo) error {
	params := url.Values{}
	params.Set("id", ipAddressId)

	resp, err := exo.AsyncRequest("removeIpFromNic", params, async)
	if err != nil {
		return err
	}

	var r BooleanResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return err
	}

	if !r.Success {
		return fmt.Errorf("Cannot removeIpFromNic. %s", r.DisplayText)
	}

	return nil
}
