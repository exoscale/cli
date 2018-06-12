package egoscale

import (
	"fmt"
	"net/url"
)

// ListRequest builds the ListNetworks request
func (network *Network) ListRequest() (ListCommand, error) {
	//TODO add tags support
	req := &ListNetworks{
		Account:           network.Account,
		ACLType:           network.ACLType,
		CanUseForDeploy:   &network.CanUseForDeploy,
		DomainID:          network.DomainID,
		ID:                network.ID,
		PhysicalNetworkID: network.PhysicalNetworkID,
		RestartRequired:   &network.RestartRequired,
		TrafficType:       network.TrafficType,
		Type:              network.Type,
		ZoneID:            network.ZoneID,
	}

	return req, nil
}

// ResourceType returns the type of the resource
func (*Network) ResourceType() string {
	return "Network"
}

func (*CreateNetwork) name() string {
	return "createNetwork"
}

func (*CreateNetwork) description() string {
	return "Creates a network"
}

func (*CreateNetwork) response() interface{} {
	return new(Network)
}

func (req *CreateNetwork) onBeforeSend(params *url.Values) error {
	// Those fields are required but might be empty
	if req.Name == "" {
		params.Set("name", "")
	}
	if req.DisplayText == "" {
		params.Set("displaytext", "")
	}
	return nil
}

func (*UpdateNetwork) name() string {
	return "updateNetwork"
}

func (*UpdateNetwork) description() string {
	return "Updates a network"
}

func (*UpdateNetwork) asyncResponse() interface{} {
	return new(Network)
}

func (*RestartNetwork) name() string {
	return "restartNetwork"
}

func (*RestartNetwork) description() string {
	return "Restarts the network; includes 1) restarting network elements - virtual routers, dhcp servers 2) reapplying all public ips 3) reapplying loadBalancing/portForwarding rules"
}

func (*RestartNetwork) asyncResponse() interface{} {
	return new(Network)
}

func (*DeleteNetwork) name() string {
	return "deleteNetwork"
}

func (*DeleteNetwork) description() string {
	return "Deletes a network"
}

func (*DeleteNetwork) asyncResponse() interface{} {
	return new(booleanResponse)
}

func (*ListNetworks) name() string {
	return "listNetworks"
}

func (*ListNetworks) description() string {
	return "Lists all available networks."
}

func (*ListNetworks) response() interface{} {
	return new(ListNetworksResponse)
}

// SetPage sets the current page
func (listNetwork *ListNetworks) SetPage(page int) {
	listNetwork.Page = page
}

// SetPageSize sets the page size
func (listNetwork *ListNetworks) SetPageSize(pageSize int) {
	listNetwork.PageSize = pageSize
}

func (*ListNetworks) each(resp interface{}, callback IterateItemFunc) {
	networks, ok := resp.(*ListNetworksResponse)
	if !ok {
		callback(nil, fmt.Errorf("wrong type. ListNetworksResponse expected, got %T", resp))
		return
	}

	for i := range networks.Network {
		if !callback(&networks.Network[i], nil) {
			break
		}
	}
}
