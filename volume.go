package egoscale

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ListVolumes
func (exo *Client) ListVolumes(params url.Values) ([]*Volume, error) {
	resp, err := exo.Request("listVolumes", params)
	if err != nil {
		return nil, err
	}

	var r ListVolumesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.Volume, nil
}

// GetRootVolumeForVirtualMachine(d.Id())
func (exo *Client) GetRootVolumeForVirtualMachine(virtualMachineId string) (*Volume, error) {
	params := url.Values{}
	params.Set("virtualmachineid", virtualMachineId)
	params.Set("type", "ROOT")

	volumes, err := exo.ListVolumes(params)
	if err != nil {
		return nil, err
	}

	if len(volumes) != 1 {
		return nil, fmt.Errorf("Expected exactly one volume for %v, got %d", virtualMachineId, len(volumes))
	}

	return volumes[0], nil
}
