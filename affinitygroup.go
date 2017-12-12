package egoscale

import (
	"encoding/json"
	"net/url"
)

func (exo *Client) CreateAffinityGroup(name string, async AsyncInfo) (*AffinityGroup, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("type", "host anti-affinity")

	resp, err := exo.AsyncRequest("createAffinityGroup", params, async)
	if err != nil {
		return nil, err
	}

	var r CreateAffinityGroupResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r.AffinityGroup, nil
}

func (exo *Client) DeleteAffinityGroup(name string, async AsyncInfo) (bool, error) {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.AsyncRequest("deleteAffinityGroup", params, async)
	if err != nil {
		return false, err
	}

	var r BooleanResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return false, err
	}

	return r.Success, nil
}
