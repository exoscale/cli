package egoscale

import (
	"net/url"
	"testing"
)

func TestNetwork(t *testing.T) {
	instance := &Network{}
	if instance.ResourceType() != "Network" {
		t.Errorf("ResourceType doesn't match")
	}
}

func TestListNetworks(t *testing.T) {
	req := &ListNetworks{}
	_ = req.response().(*ListNetworksResponse)
}

func TestCreateNetwork(t *testing.T) {
	req := &CreateNetwork{}
	_ = req.response().(*Network)
}

func TestRestartNetwork(t *testing.T) {
	req := &RestartNetwork{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*Network)
}

func TestUpdateNetwork(t *testing.T) {
	req := &UpdateNetwork{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*Network)
}

func TestDeleteNetwork(t *testing.T) {
	req := &DeleteNetwork{}
	_ = req.response().(*AsyncJobResult)
	_ = req.asyncResponse().(*booleanResponse)
}

func TestCreateNetworkOnBeforeSend(t *testing.T) {
	req := &CreateNetwork{}
	params := url.Values{}

	if err := req.onBeforeSend(params); err != nil {
		t.Error(err)
	}

	if _, ok := params["name"]; !ok {
		t.Errorf("name should have been set")
	}
	if _, ok := params["displaytext"]; !ok {
		t.Errorf("displaytext should have been set")
	}
}

func TestListNetwork(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnetworksresponse": {
	"count": 5,
	"network": [
	  {
		"account": "exoscale-1",
		"acltype": "Account",
		"broadcastdomaintype": "Vxlan",
		"canusefordeploy": true,
		"displaytext": "hello",
		"domain": "exoscale-1",
		"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
		"id": "939a8c40-75b5-4a82-9d7e-f8813a26cf7c",
		"ispersistent": true,
		"issystem": false,
		"name": "testtest",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Private Network",
		"networkofferingid": "eb35f4e6-0ecc-412e-9925-e469bf03d8fd",
		"networkofferingname": "PrivNet",
		"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
		"related": "939a8c40-75b5-4a82-9d7e-f8813a26cf7c",
		"restartrequired": false,
		"service": [
		  {
			"name": "PrivateNetwork"
		  }
		],
		"specifyipranges": false,
		"state": "Implemented",
		"strechedl2subnet": false,
		"tags": [],
		"traffictype": "Guest",
		"type": "Isolated",
		"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
		"zonename": "ch-dk-2"
	  },
	  {
		"acltype": "Domain",
		"broadcastdomaintype": "Vlan",
		"canusefordeploy": true,
		"displaytext": "defaultGuestNetwork",
		"dns1": "8.8.8.8",
		"dns2": "8.8.4.4",
		"domain": "ROOT",
		"domainid": "4a8857b8-7235-4e31-a7ef-b8b44d180850",
		"id": "bf25cad3-e0fa-4838-80e0-4b20e4bfb245",
		"ip6cidr": "2a04:c46:e00::/40",
		"ispersistent": false,
		"issystem": false,
		"name": "defaultGuestNetwork",
		"networkdomain": "cs1cloud.internal",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Offering for Shared Security group enabled networks",
		"networkofferingid": "7d458d69-0eae-42ab-9973-8d119160e3ca",
		"networkofferingname": "DefaultSharedNetworkOfferingWithSGService",
		"physicalnetworkid": "10101734-124f-4aae-a8ac-2b36e216cf75",
		"related": "bf25cad3-e0fa-4838-80e0-4b20e4bfb245",
		"restartrequired": false,
		"service": [
		  {
			"name": "SecurityGroup"
		  },
		  {
			"name": "UserData"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "DhcpAccrossMultipleSubnets",
				"value": "true"
			  }
			],
			"name": "Dhcp"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "AllowDnsSuffixModification",
				"value": "true"
			  }
			],
			"name": "Dns"
		  }
		],
		"specifyipranges": true,
		"state": "Setup",
		"strechedl2subnet": false,
		"subdomainaccess": true,
		"tags": [],
		"traffictype": "Guest",
		"type": "Shared",
		"zoneid": "35eb7739-d19e-45f7-a581-4687c54d6d02",
		"zonename": "de-fra-1"
	  },
	  {
		"acltype": "Domain",
		"broadcastdomaintype": "Vlan",
		"canusefordeploy": true,
		"displaytext": "defaultGuestNetwork",
		"dns1": "8.8.8.8",
		"dns2": "8.8.4.4",
		"domain": "ROOT",
		"domainid": "4a8857b8-7235-4e31-a7ef-b8b44d180850",
		"id": "05c0e278-7ab4-4a6d-aa9c-3158620b6471",
		"ip6cidr": "2a04:c45:e00::/40",
		"ispersistent": false,
		"issystem": false,
		"name": "defaultGuestNetwork",
		"networkdomain": "cs1cloud.internal",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Offering for Shared Security group enabled networks",
		"networkofferingid": "7d458d69-0eae-42ab-9973-8d119160e3ca",
		"networkofferingname": "DefaultSharedNetworkOfferingWithSGService",
		"physicalnetworkid": "a2964a38-f6a0-454c-b2a0-82bb50b433e4",
		"related": "05c0e278-7ab4-4a6d-aa9c-3158620b6471",
		"restartrequired": false,
		"service": [
		  {
			"name": "SecurityGroup"
		  },
		  {
			"name": "UserData"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "DhcpAccrossMultipleSubnets",
				"value": "true"
			  }
			],
			"name": "Dhcp"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "AllowDnsSuffixModification",
				"value": "true"
			  }
			],
			"name": "Dns"
		  }
		],
		"specifyipranges": true,
		"state": "Setup",
		"strechedl2subnet": false,
		"subdomainaccess": true,
		"tags": [],
		"traffictype": "Guest",
		"type": "Shared",
		"zoneid": "4da1b188-dcd6-4ff5-b7fd-bde984055548",
		"zonename": "at-vie-1"
	  },
	  {
		"acltype": "Domain",
		"broadcastdomaintype": "Native",
		"canusefordeploy": true,
		"displaytext": "defaultGuestNetwork",
		"dns1": "8.8.8.8",
		"dns2": "8.8.4.4",
		"domain": "ROOT",
		"domainid": "4a8857b8-7235-4e31-a7ef-b8b44d180850",
		"id": "e38bf934-1277-4e82-8cef-f3a862d9ec57",
		"ip6cidr": "2a04:c44:e00::/40",
		"ispersistent": false,
		"issystem": false,
		"name": "defaultGuestNetwork",
		"networkdomain": "cs1cloud.internal",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Offering for Shared Security group enabled networks",
		"networkofferingid": "7d458d69-0eae-42ab-9973-8d119160e3ca",
		"networkofferingname": "DefaultSharedNetworkOfferingWithSGService",
		"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
		"related": "e38bf934-1277-4e82-8cef-f3a862d9ec57",
		"restartrequired": false,
		"service": [
		  {
			"name": "SecurityGroup"
		  },
		  {
			"name": "UserData"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "DhcpAccrossMultipleSubnets",
				"value": "true"
			  }
			],
			"name": "Dhcp"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "AllowDnsSuffixModification",
				"value": "true"
			  }
			],
			"name": "Dns"
		  }
		],
		"specifyipranges": true,
		"state": "Setup",
		"strechedl2subnet": false,
		"subdomainaccess": true,
		"tags": [],
		"traffictype": "Guest",
		"type": "Shared",
		"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
		"zonename": "ch-dk-2"
	  },
	  {
		"acltype": "Domain",
		"broadcastdomaintype": "Vlan",
		"canusefordeploy": true,
		"displaytext": "guestNetworkForBasicZone",
		"dns1": "8.8.8.8",
		"dns2": "8.8.4.4",
		"domain": "ROOT",
		"domainid": "4a8857b8-7235-4e31-a7ef-b8b44d180850",
		"gateway": "185.19.28.1",
		"id": "00304a04-c7ea-4e77-a786-18bc64347bf7",
		"ip6cidr": "2a04:c43:e00::/40",
		"ispersistent": false,
		"issystem": false,
		"name": "guestNetworkForBasicZone",
		"networkdomain": "cs1cloud.internal",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Offering for Shared Security group enabled networks",
		"networkofferingid": "7d458d69-0eae-42ab-9973-8d119160e3ca",
		"networkofferingname": "DefaultSharedNetworkOfferingWithSGService",
		"physicalnetworkid": "07f747f5-b445-487f-b2d7-81a5a512989e",
		"related": "00304a04-c7ea-4e77-a786-18bc64347bf7",
		"restartrequired": false,
		"service": [
		  {
			"name": "SecurityGroup"
		  },
		  {
			"name": "UserData"
		  },
		  {
			"capability": [
			  {
				"canchooseservicecapability": false,
				"name": "DhcpAccrossMultipleSubnets",
				"value": "true"
			  }
			],
			"name": "Dhcp"
		  }
		],
		"specifyipranges": true,
		"state": "Setup",
		"strechedl2subnet": false,
		"subdomainaccess": true,
		"tags": [],
		"traffictype": "Guest",
		"type": "Shared",
		"zoneid": "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		"zonename": "ch-gva-2"
	  }
	]
  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	network := new(Network)
	networks, err := cs.List(network)
	if err != nil {
		t.Fatal(err)
	}

	if len(networks) != 5 {
		t.Errorf("Five networks were expected, got %d", len(networks))
	}

	if networks[0].(*Network).ID != "939a8c40-75b5-4a82-9d7e-f8813a26cf7c" {
		t.Errorf("Expected ID, got %#v", networks[0])
	}
}

func TestListNetworkEmpty(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnetworksresponse": {
	"count": 0,
	"network": []
  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	network := new(Network)
	networks, err := cs.List(network)
	if err != nil {
		t.Fatal(err)
	}

	if len(networks) != 0 {
		t.Errorf("zero networks were expected, got %d", len(networks))
	}
}

func TestListNetworkFailure(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnetworksresponse": {
	"count": 3456,
	"network": {}
  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	network := new(Network)
	networks, err := cs.List(network)
	if err == nil {
		t.Errorf("error was expected, got %v", err)
	}

	if len(networks) != 0 {
		t.Errorf("zero networks were expected, got %d", len(networks))
	}
}

func TestListNetworkPaginate(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnetworksresponse": {
	"count": 2,
	"network": [
	  {
		"account": "exoscale-1",
		"acltype": "Account",
		"broadcastdomaintype": "Vxlan",
		"canusefordeploy": true,
		"displaytext": "fddd",
		"domain": "exoscale-1",
		"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
		"id": "772cee7a-631f-4c0e-ad2b-27776f260d71",
		"ispersistent": true,
		"issystem": false,
		"name": "klmfsdvds",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Private Network",
		"networkofferingid": "eb35f4e6-0ecc-412e-9925-e469bf03d8fd",
		"networkofferingname": "PrivNet",
		"physicalnetworkid": "07f747f5-b445-487f-b2d7-81a5a512989e",
		"related": "772cee7a-631f-4c0e-ad2b-27776f260d71",
		"restartrequired": false,
		"service": [
		  {
			"name": "PrivateNetwork"
		  }
		],
		"specifyipranges": false,
		"state": "Implemented",
		"strechedl2subnet": false,
		"tags": [],
		"traffictype": "Guest",
		"type": "Isolated",
		"zoneid": "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		"zonename": "ch-gva-2"
	  },
	  {
		"account": "exoscale-1",
		"acltype": "Account",
		"broadcastdomaintype": "Vxlan",
		"canusefordeploy": true,
		"displaytext": "sss",
		"domain": "exoscale-1",
		"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
		"id": "73c2c6a3-c3b6-447c-a50f-5443bc74cfd2",
		"ispersistent": true,
		"issystem": false,
		"name": "test",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Private Network",
		"networkofferingid": "eb35f4e6-0ecc-412e-9925-e469bf03d8fd",
		"networkofferingname": "PrivNet",
		"physicalnetworkid": "0e2989e9-4cac-4faa-a6c6-8e43f936a76a",
		"related": "73c2c6a3-c3b6-447c-a50f-5443bc74cfd2",
		"restartrequired": false,
		"service": [
		  {
			"name": "PrivateNetwork"
		  }
		],
		"specifyipranges": false,
		"state": "Implemented",
		"strechedl2subnet": false,
		"tags": [],
		"traffictype": "Guest",
		"type": "Isolated",
		"zoneid": "91e5e9e4-c9ed-4b76-bee4-427004b3baf9",
		"zonename": "ch-dk-2"
	  }
	]
  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	network := &Network{CanUseForDeploy: true, Type: "Isolated"}

	req, err := network.ListRequest()
	if err != nil {
		t.Fatal(err)
	}
	cs.Paginate(req, func(i interface{}, err error) bool {
		if err != nil {
			t.Fatal(err)
		}

		if net := i.(*Network); net.Name != "klmfsdvds" {
			t.Errorf("klmfsdvds zone name was expected, got %s", i.(*Network).Name)
		}
		return false
	})
}

func TestFindNetwork(t *testing.T) {
	ts := newServer(response{200, jsonContentType, `
{"listnetworksresponse": {
	"count": 1,
	"network": [
	  {
		"account": "exoscale-1",
		"acltype": "Account",
		"broadcastdomaintype": "Vxlan",
		"canusefordeploy": true,
		"displaytext": "fddd",
		"domain": "exoscale-1",
		"domainid": "5b2f621e-3eb6-4a14-a315-d4d7d62f28ff",
		"id": "772cee7a-631f-4c0e-ad2b-27776f260d71",
		"ispersistent": true,
		"issystem": false,
		"name": "klmfsdvds",
		"networkofferingavailability": "Optional",
		"networkofferingconservemode": true,
		"networkofferingdisplaytext": "Private Network",
		"networkofferingid": "eb35f4e6-0ecc-412e-9925-e469bf03d8fd",
		"networkofferingname": "PrivNet",
		"physicalnetworkid": "07f747f5-b445-487f-b2d7-81a5a512989e",
		"related": "772cee7a-631f-4c0e-ad2b-27776f260d71",
		"restartrequired": false,
		"service": [
		  {
			"name": "PrivateNetwork"
		  }
		],
		"specifyipranges": false,
		"state": "Implemented",
		"strechedl2subnet": false,
		"tags": [],
		"traffictype": "Guest",
		"type": "Isolated",
		"zoneid": "1128bd56-b4d9-4ac6-a7b9-c715b187ce11",
		"zonename": "ch-gva-2"
	  }
	]
  }}`})
	defer ts.Close()

	cs := NewClient(ts.URL, "KEY", "SECRET")

	networks, err := cs.List(&Network{Name: "klmfsdvds", CanUseForDeploy: true, Type: "Isolated"})
	if err != nil {
		t.Fatal(err)
	}

	if len(networks) != 1 {
		t.Fatalf("One network was expected, got %d", len(networks))
	}

	net, ok := networks[0].(*Network)
	if !ok {
		t.Errorf("unable to type inference *Network, got %v", net)
	}

	if networks[0].(*Network).Name != "klmfsdvds" {
		t.Errorf("klmfsdvds network name was expected, got %s", networks[0].(*Network).Name)
	}

}
