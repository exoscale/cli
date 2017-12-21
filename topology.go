package egoscale

import (
	"fmt"
	"regexp"
	"strings"
)

// Deprecated: GetSecurityGroups returns all security groups
func (exo *Client) GetSecurityGroups() (map[string]SecurityGroup, error) {
	var sgs map[string]SecurityGroup
	resp := new(ListSecurityGroupsResponse)
	err := exo.Request(&ListSecurityGroupsRequest{}, resp)
	if err != nil {
		return nil, err
	}

	sgs = make(map[string]SecurityGroup)
	for _, sg := range resp.SecurityGroup {
		sgs[sg.Name] = *sg
	}
	return sgs, nil
}

// Deprecated: GetSecurityGroupId returns security group by name
func (exo *Client) GetSecurityGroupId(name string) (string, error) {
	resp := new(ListSecurityGroupsResponse)
	err := exo.Request(&ListSecurityGroupsRequest{SecurityGroupName: name}, resp)
	if err != nil {
		return "", err
	}

	for _, sg := range resp.SecurityGroup {
		if sg.Name == name {
			return sg.Id, nil
		}
	}

	return "", nil
}

// Deprecated: GetAllZones returns all the zone id by name
func (exo *Client) GetAllZones() (map[string]string, error) {
	var zones map[string]string
	r := new(ListZonesResponse)
	err := exo.Request(&ListZonesRequest{}, r)
	if err != nil {
		return zones, err
	}

	zones = make(map[string]string)
	for _, zone := range r.Zone {
		zones[strings.ToLower(zone.Name)] = zone.Id
	}
	return zones, nil
}

// Deprecated: GetProfiles returns a mapping of the service offerings by name
func (exo *Client) GetProfiles() (map[string]string, error) {
	profiles := make(map[string]string)
	r := new(ListServiceOfferingsResponse)
	err := exo.Request(&ListServiceOfferingsRequest{}, r)
	if err != nil {
		return profiles, nil
	}

	for _, offering := range r.ServiceOffering {
		profiles[strings.ToLower(offering.Name)] = offering.Id
	}

	return profiles, nil
}

// Deprecated: GetKeypairs returns the list of SSH keyPairs
func (exo *Client) GetKeypairs() ([]SshKeyPair, error) {
	var keypairs []SshKeyPair

	resp := new(ListSshKeyPairsResponse)
	err := exo.Request(&ListSshKeyPairsRequest{}, resp)
	if err != nil {
		return keypairs, err
	}
	keypairs = make([]SshKeyPair, resp.Count)
	for i, keypair := range resp.SshKeyPair {
		keypairs[i] = *keypair
	}
	return keypairs, nil
}

func (exo *Client) GetAffinityGroups() (map[string]string, error) {
	var affinitygroups map[string]string

	resp := new(ListAffinityGroupsResponse)
	err := exo.Request(&ListAffinityGroupsRequest{}, resp)
	if err != nil {
		return affinitygroups, err
	}

	affinitygroups = make(map[string]string)
	for _, affinitygroup := range resp.AffinityGroup {
		affinitygroups[affinitygroup.Name] = affinitygroup.Id
	}
	return affinitygroups, nil
}

// Deprecated: GetImages list the available featured images and group them by name, then size.
func (exo *Client) GetImages() (map[string]map[int64]string, error) {
	var images map[string]map[int64]string
	images = make(map[string]map[int64]string)
	re := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9.]+)\b`)

	resp := new(ListTemplatesResponse)
	err := exo.Request(&ListTemplatesRequest{
		TemplateFilter: "featured",
		ZoneId:         "1", // XXX: Hack to list only CH-GVA
	}, resp)
	if err != nil {
		return images, err
	}

	for _, template := range resp.Template {
		size := int64(template.Size >> 30) // B to GiB

		fullname := strings.ToLower(template.Name)

		if _, present := images[fullname]; !present {
			images[fullname] = make(map[int64]string)
		}
		images[fullname][size] = template.Id

		submatch := re.FindStringSubmatch(template.Name)
		if len(submatch) > 0 {
			name := strings.Replace(strings.ToLower(submatch[1]), " ", "-", -1)
			version := submatch[2]
			image := fmt.Sprintf("%s-%s", name, version)

			if _, present := images[image]; !present {
				images[image] = make(map[int64]string)
			}
			images[image][size] = template.Id
		}
	}
	return images, nil
}

// Deprecated: GetTopology returns an big, yet incomplete view of the world
func (exo *Client) GetTopology() (*Topology, error) {
	zones, err := exo.GetAllZones()
	if err != nil {
		return nil, err
	}
	images, err := exo.GetImages()
	if err != nil {
		return nil, err
	}
	securityGroups, err := exo.GetSecurityGroups()
	if err != nil {
		return nil, err
	}
	groups := make(map[string]string)
	for k, v := range securityGroups {
		groups[k] = v.Id
	}

	keypairs, err := exo.GetKeypairs()
	if err != nil {
		return nil, err
	}

	/* Convert the ssh keypair to contain just the name */
	keynames := make([]string, len(keypairs))
	for i, k := range keypairs {
		keynames[i] = k.Name
	}

	affinitygroups, err := exo.GetAffinityGroups()
	if err != nil {
		return nil, err
	}

	profiles, err := exo.GetProfiles()
	if err != nil {
		return nil, err
	}

	topo := &Topology{
		Zones:          zones,
		Images:         images,
		Keypairs:       keynames,
		Profiles:       profiles,
		AffinityGroups: affinitygroups,
		SecurityGroups: groups,
	}

	return topo, nil
}
