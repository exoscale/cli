package egoscale

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetImages(t *testing.T) {
	ts := newServer(`
{
	"listtemplatesresponse (doesn't matter)": {
		"count": 0,
		"template": [
			{
				"id": "4c0732a0-3df0-4f66-8d16-009f91cf05d6",
				"name": "Linux RedHat 7.4 64-bit",
				"displayText": "Linux RedHat 7.4 64-bit 10G Disk (2017-11-31-dummy)",
				"size": 10737418240
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux Ubuntu 12.04 LTS 64-bit",
				"displayText": "Linux Ubuntu 12.04 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux Debian 8 64-bit",
				"displayText": "Linux Debian 8 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux CentOS 7.3 64-bit",
				"displayText": "Linux CentOS 7.3 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux CoreOS stable 1298 64-bit",
				"displayText": "Linux CoreOS stable 1298 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			}
		]
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	images, err := cs.GetImages()
	if err != nil {
		log.Fatal(err)
	}

	var tests = []struct {
		uuid  string
		names []string
		size  int
	}{
		{
			"4c0732a0-3df0-4f66-8d16-009f91cf05d6",
			[]string{"redhat-7.4", "linux redhat 7.4 64-bit"},
			10,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"ubuntu-12.04", "linux ubuntu 12.04 lts 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"debian-8", "linux debian 8 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"centos-7.3", "linux centos 7.3 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"coreos-stable-1298", "linux coreos stable 1298 64-bit"},
			50,
		},
	}

	for _, test := range tests {
		for _, name := range test.names {
			if _, ok := images[name]; !ok {
				t.Errorf("expected %s into the map", name)
			}

			if _, ok := images[name][test.size]; !ok {
				t.Errorf("expected %s, %dG into the map", name, test.size)
			}

			if uuid := images[name][test.size]; uuid != test.uuid {
				t.Errorf("bad uuid for the %s image. got %v expected %v", name, uuid, test.uuid)
			}
		}
	}
}

func TestGetSecurityGroups(t *testing.T) {
	ts := newServer(`
{
	"listsecurityresponse (doesn't matter)": {
		"count": 1,
		"securitygroup": [
			{
				"account": "john.doe@example.org",
				"description": "Default Security Group",
				"egressrule": [],
				"id": "8282c50e-db68-4584-84ef-394ca68165fc",
				"ingressrule": [
					{
						"cidr": "0.0.0.0/0",
						"endport": 22,
						"protocol": "tcp",
						"ruleid": "933aa3f0-1e0b-4428-ab13-ee0bd0874f03",
						"startport": 22,
						"tags": []
					},
					{
						"cidr": "0.0.0.0/0",
						"icmpcode": 0,
						"icmptype": 8,
						"protocol": "icmp",
						"ruleid": "db864f8d-6f08-4fa6-84e1-ba1742930db6",
						"tags": []
					},
					{
						"protocol": "tcp",
						"startport": 80,
						"endport": 80,
						"usersecuritygrouplist": [
							{
								"account": "john.doe@example.org",
								"group": "other"
							}
						]
					}
				],
				"name": "dummy",
				"tags": []
			}
		]
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	params := url.Values{}
	securityGroups, err := cs.GetSecurityGroups(params)
	if err != nil {
		log.Fatal(err)
	}

	sg := securityGroups[0]
	if sg.IngressRules[2].UserSecurityGroupList[0].Group != "other" {
		t.Errorf("UserSecurityGroupList %s not found", "other")
	}
}

func newServer(response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}
