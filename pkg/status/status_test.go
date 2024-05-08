package status_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exoscale/cli/pkg/status"
)

// Status page and the services
var svc1Parent = "Parent"
var svc1Child1 = "API"
var svc1Child2 = "Web app"

var svc2Parent = "NoChild"
var svc3Parent = "OtherParent"
var svc3Child1 = "API"

// Produce JSON with services and with or without events
func getJSONStatus(events bool) string {

	incidentType := "null"
	maintenanceType := "null"
	incidents := `[]`
	maintenances := `[]`

	// Add maintenance and incident
	if events {
		// Update incident type
		incidentType = "\"minor\""
		maintenanceType = "\"scheduled\""

		incidents = `[
      {"id":123,
       "service_ids":[10,11,30,31],
       "starts_at":"2023-07-28T07:57:37",
       "title":"App partial outage",
       "type":"minor"
      }]`
		maintenances = `[
      {"id":124,
       "service_ids":[10,12],
       "starts_at":"2023-07-28T11:00:00",
       "title":"app maintenance",
       "type":"scheduled"
      } ]`
	}

	// build Json
	return fmt.Sprintf(`
      {
        "services": [
          {"name": "%v",
           "id": 10,
           "current_incident_type": %v,
           "children": [
            {"name": "%v",
             "id": 11,
             "current_incident_type": %v},
            {"name": "%v",
             "id": 12,
             "current_incident_type": %v}]},
            {"name": "%v",
             "id": 20,
             "current_incident_type": null,
             "children": null},
            {"name": "%v",
             "id": 30,
             "current_incident_type": %v,
             "children": [
              {"name": "%v",
               "id": 31,
               "current_incident_type": %v}]}
        ],
        "incidents": %v,
        "maintenances": %v
      }`, svc1Parent, incidentType, svc1Child1, incidentType, svc1Child2, maintenanceType,
		svc2Parent, svc3Parent, incidentType, svc3Child1, incidentType,
		incidents, maintenances)
}

// Test status page
func TestStatusPageNoEvent(t *testing.T) {

	var statusPageTest status.StatusPalStatus

	// Json without incidents or maintenances
	jsonData := getJSONStatus(false)
	err := json.Unmarshal([]byte(jsonData), &statusPageTest)
	assert.NoError(t, err, err)

	// Validate the services returned by GetStatusByZone
	output, err := statusPageTest.GetStatusByZone()
	//	assert.Equal(t, 3, len(output), jsonData)
	// expected Services are only "parent" services
	expectedServices := []string{svc1Parent, svc2Parent, svc3Parent}
	assert.Equal(t, len(expectedServices), len(output), "Only parent services are expected")
	assert.NoError(t, err)

	// check the parent service names
	for i, o := range output {
		assert.Equal(t, o[0], expectedServices[i], "Service name expected")
		fmt.Print(o)
	}

	i, err := statusPageTest.Incidents.GetActiveEvents(statusPageTest.Services)
	assert.Nil(t, i, "No Incident expected")
	assert.NoError(t, err)
	m, err := statusPageTest.Maintenances.GetActiveEvents(statusPageTest.Services)
	assert.Nil(t, m, "No Maintenance expected")
	assert.NoError(t, err)

	// Check IsParentServce
	assert.Equal(t, true, statusPageTest.Services.IsParentService(10))
	assert.Equal(t, false, statusPageTest.Services.IsParentService(11))

	// Check ServiceName contains Parent
	name, err := statusPageTest.Services.GetServiceNamebyID(10)
	assert.NoError(t, err)
	assert.Equal(t, svc1Parent, name)
	name, err = statusPageTest.Services.GetServiceNamebyID(11)
	assert.NoError(t, err)
	assert.Equal(t, svc1Parent+" "+svc1Child1, name)

}

// Test status page with active incident and maintenance
func TestStatusPageEvents(t *testing.T) {

	var statusPageTest status.StatusPalStatus

	// Json without incidents or maintenances
	jsonData := getJSONStatus(true)
	err := json.Unmarshal([]byte(jsonData), &statusPageTest)
	assert.NoError(t, err, err)

	i, err := statusPageTest.Incidents.GetActiveEvents(statusPageTest.Services)
	assert.NotNil(t, i, "Incident expected")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(i), "2 (child) services impacted by the incident")
	assert.Equal(t, 4, len(i[0]), "4 fields expected: Name, title, type and start time")
	assert.Equal(t, 4, len(i[1]), "4 fields expected: Name, title, type and start time")
	// Name = full service name and result is sorted by Name
	// it should be svc3 first
	name, _ := statusPageTest.Services.GetServiceNamebyID(31)
	assert.Equal(t, name, i[0][0], "Incidents should be sorted by name")
	assert.Equal(t, "minor", i[0][2], "Incident Type minor")

	m, err := statusPageTest.Maintenances.GetActiveEvents(statusPageTest.Services)
	assert.NotNil(t, m, "Maintenance expected")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m), "1 (child) service impacted by the maintenance")
	assert.Equal(t, 3, len(m[0]), "3 fields expected: Name, title and start time")

}
