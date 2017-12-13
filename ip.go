package egoscale

import (
	"encoding/json"
	"fmt"
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
func (exo *Client) DestroyIpAddress(ipAddressId string, async AsyncInfo) error {
	return exo.DisassociateIpAddress(ipAddressId, async)
}

// DisassociateIpAddress disassociates a public IP from the account
func (exo *Client) DisassociateIpAddress(ipAddressId string, async AsyncInfo) error {
	params := url.Values{}
	params.Set("id", ipAddressId)
	resp, err := exo.AsyncRequest("disassociateIpAddress", params, async)
	if err != nil {
		return err
	}

	var r BooleanResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return err
	}

	if !r.Success {
		return fmt.Errorf("Cannot disassociateIpAddress. %s", r.DisplayText)
	}

	return nil
}

// GetAllIpAddresses returns all the public IP addresses
func (exo *Client) GetAllIpAddresses() ([]*IpAddress, error) {
	params := url.Values{}

	return exo.ListPublicIpAddresses(params)
}

// ListPublicIpAddresses lists the public ip addresses
func (exo *Client) ListPublicIpAddresses(params url.Values) ([]*IpAddress, error) {
	resp, err := exo.Request("listPublicIpAddresses", params)
	if err != nil {
		return nil, err
	}

	var r ListPublicIpAddressesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.PublicIpAddress, nil
}
