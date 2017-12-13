package egoscale

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// CreateVirtualMachine is an alias for DeployVirtualMachine
func (exo *Client) CreateVirtualMachine(p VirtualMachineProfile, async AsyncInfo) (*VirtualMachine, error) {
	return exo.DeployVirtualMachine(p, async)
}

// DeployVirtualMachine creates a new VM
func (exo *Client) DeployVirtualMachine(p VirtualMachineProfile, async AsyncInfo) (*VirtualMachine, error) {
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
	if len(p.AffinityGroups) > 0 {
		params.Set("affinitygroupnames", strings.Join(p.AffinityGroups, ","))
	}

	params.Set("securitygroupids", strings.Join(p.SecurityGroups, ","))

	return exo.doVirtualMachine("deploy", params, async)
}

// StartVirtualMachine starts the VM and returns its new state
func (exo *Client) StartVirtualMachine(virtualMachineId string, async AsyncInfo) (*VirtualMachine, error) {
	params := url.Values{}
	params.Set("id", virtualMachineId)

	return exo.doVirtualMachine("start", params, async)
}

// StopVirtualMachine stops the VM and returns its new state
func (exo *Client) StopVirtualMachine(virtualMachineId string, async AsyncInfo) (*VirtualMachine, error) {
	params := url.Values{}
	params.Set("id", virtualMachineId)

	return exo.doVirtualMachine("stop", params, async)
}

// RebootVirtualMachine reboots the VM and returns its new state
func (exo *Client) RebootVirtualMachine(virtualMachineId string, async AsyncInfo) (*VirtualMachine, error) {
	params := url.Values{}
	params.Set("id", virtualMachineId)

	return exo.doVirtualMachine("reboot", params, async)
}

// DeleteVirtualMachine is an alias for DestroyVirtualMachine
func (exo *Client) DeleteVirtualMachine(virtualMachineId string, async AsyncInfo) (*VirtualMachine, error) {
	return exo.DestroyVirtualMachine(virtualMachineId, async)
}

// DestroyVirtualMachine destroy the VM
func (exo *Client) DestroyVirtualMachine(virtualMachineId string, async AsyncInfo) (*VirtualMachine, error) {
	params := url.Values{}
	params.Set("id", virtualMachineId)

	return exo.doVirtualMachine("destroy", params, async)
}

// doVirtualMachine is a utility function to perform the API call
func (exo *Client) doVirtualMachine(action string, params url.Values, async AsyncInfo) (*VirtualMachine, error) {
	resp, err := exo.AsyncRequest(action+"VirtualMachine", params, async)
	if err != nil {
		return nil, err
	}

	var r VirtualMachineResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.VirtualMachine, nil
}

// GetVirtualMachine
func (exo *Client) GetVirtualMachine(virtualMachineId string) (*VirtualMachine, error) {

	params := url.Values{}
	params.Set("id", virtualMachineId)

	resp, err := exo.Request("listVirtualMachines", params)
	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if len(r.VirtualMachines) == 1 {
		machine := r.VirtualMachines[0]
		return machine, nil
	} else {
		return nil, fmt.Errorf("cannot retrieve virtualmachine with id %s", virtualMachineId)
	}
}

// ListVirtualMachines lists all the VM
func (exo *Client) ListVirtualMachines() ([]*VirtualMachine, error) {

	resp, err := exo.Request("listVirtualMachines", url.Values{})
	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.VirtualMachines, nil
}
