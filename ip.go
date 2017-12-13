package egoscale

import (
	"encoding/json"
	"net/url"
)

// CreateIpAddress is an alias for AssociateIpAddress
func (exo *Client) CreateIpAddress(profile IpAddressProfile, async AsyncInfo) (*IpAddress, error) {
	return exo.AssociateIpAddress(profile, async)
}

// AssociateIpAddress acquires and associates a public IP to a given zone
func (exo *Client) AssociateIpAddress(profile IpAddressProfile, async AsyncInfo) (*IpAddress, error) {
	params := url.Values{}
	params.Set("zoneid", profile.Zone)
	resp, err := exo.AsyncRequest("associateIpAddress", params, async)
	if err != nil {
		return nil, err
	}

	var r AssociateIpAddressResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.IpAddress, nil
}

// DestroyIpAddress is an alias for DisassociateIpAddress
func (exo *Client) DestroyIpAddress(ipAddressId string, async AsyncInfo) (bool, error) {
	return exo.DisassociateIpAddress(ipAddressId, async)
}

// DisassociateIpAddress disassociates a public IP from the account
func (exo *Client) DisassociateIpAddress(ipAddressId string, async AsyncInfo) (bool, error) {
	params := url.Values{}
	params.Set("id", ipAddressId)
	resp, err := exo.AsyncRequest("disassociateIpAddress", params, async)
	if err != nil {
		return false, err
	}

	var r BooleanResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return false, err
	}

	return r.Success, nil
}

// AddIpToNic
func (exo *Client) AddIpToNic(nic_id string, ip_address string, async AsyncInfo) (string, error) {
	params := url.Values{}
	params.Set("nicid", nic_id)
	params.Set("ipaddress", ip_address)

	resp, err := exo.AsyncRequest("addIpToNic", params, async)
	if err != nil {
		return "", err
	}

	var r AddIpToNicResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.Id, nil
}

// RemoveIpFromNic removes the IP from the NIC
func (exo *Client) RemoveIpFromNic(nic_id string, async AsyncInfo) (bool, error) {
	params := url.Values{}
	params.Set("id", nic_id)

	resp, err := exo.AsyncRequest("removeIpFromNic", params, async)
	if err != nil {
		return false, err
	}

	var r BooleanResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return false, err
	}
	return r.Success, nil
}
