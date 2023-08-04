package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// https://www.statuspal.io/api-docs/v2
const (
	statusPalURL          = "https://statuspal.eu/api/v2/status_pages/"
	statusContentPage     = "application/json; charset=utf-8"
	dateLayout            = "2006-01-02T15:04:05"
	IncidentTypeScheduled = "scheduled"
)

// https://www.statuspal.io/api-docs#tag/Status/operation/getStatusPageStatus

// Exoscale Services: Parent / child services
// Parent services are
type StatusPalStatus struct {
	// Services: all the services of the StatusPage with the current incident type
	Services Services `json:"services"`

	// Active Incidents and Maintenances
	Incidents    Events `json:"incidents"`
	Maintenances Events `json:"maintenances"`
}

// Get the status of global services (status of global or zones, no details of impacted services)
func (s StatusPalStatus) GetStatusByZone() ([][]string, error) {
	global := make([][]string, len(s.Services))
	for index, svc := range s.Services {
		state := svc.getIncidentType()
		global[index] = []string{*svc.Name, state}
	}
	return global, nil
}

// A service can contains several child services
// In our case:
// - Parent services = Global and all the zones
// - Child services = products available in a zone or globally
type Service struct {
	Id   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`

	// The type of the current incident:
	//  * `major` - A minor incident is currently taking place.
	//  * `minor` - A major incident is currently taking place.
	//  * `scheduled` - A scheduled maintenance is currently taking place.
	//  * null - No incident is taking place.
	IncidentType *string `json:"current_incident_type,omitempty"`

	// Each product available in the zone
	Children Services `json:"children,omitempty"`
}

func (s *Service) getIncidentType() string {
	if s.IncidentType == nil {
		return "operational"
	}
	switch *s.IncidentType {
	case IncidentTypeScheduled:
		return "scheduled maintenance"
	default:
		return fmt.Sprint(*s.IncidentType)
	}
}

// Active Maintenance or Incident
type Event struct {
	Id    *int    `json:"id,omitempty"`
	Title *string `json:"title,omitempty"`
	// The time at which the incident/maintenance started(UTC).
	StartsAt *string `json:"starts_at"`
	// Type of current incident (major, minor, scheduled)

	Type *string `json:"type"`

	// Services impacted
	ServiceIds []int `json:"service_ids"`
}

type Events []Event

// Get Active Incidents or Maintenances with the full service name (Zone+Product)
func (e Events) GetActiveEvents(services Services) ([][]string, error) {
	var events EventsDetails

	for _, event := range e {
		// Get all the services impacted by the incident (name and id)
		// Child and parent are all mixed, we need to rebuild the dependency
		for _, impacted := range event.ServiceIds {
			if services.IsParentService(impacted) {
				continue
			}
			svcName, err := services.GetServiceNamebyId(impacted)
			if err != nil {
				return nil, err
			}
			started, err := time.Parse(dateLayout, *event.StartsAt)
			if err != nil {
				return nil, err
			}
			startTimeUTC := started.Format(time.RFC822)
			eventDetails := []string{svcName, *event.Title}
			if *event.Type == IncidentTypeScheduled {
				eventDetails = append(eventDetails, "scheduled at "+startTimeUTC)
				events = append(events, eventDetails)
			} else {

				eventDetails = append(eventDetails, fmt.Sprint(*event.Type), "since "+startTimeUTC)
				events = append(events, eventDetails)
			}
		}
	}
	// Sort by zones
	sort.Sort(events)
	return events, nil
}

type Services []Service

// We have 2 levels of services, check if a service is a parent
func (s Services) IsParentService(id int) bool {
	for _, service := range s {
		if service.Id != nil && *service.Id == id {
			return true
		}
	}
	return false

}

// Return the Zone and the impacted service = fullname (parent svc + child svc)
func (s Services) GetServiceNamebyId(id int) (string, error) {
	// For all zones / global services
	for _, parentService := range s {
		// id provided is a parent service, return the name
		if *parentService.Id == id {
			return *parentService.Name, nil
		}
		// Try to find the Service Id in the child services
		for _, childService := range parentService.Children {
			// In this case, we returen the Parent and the Child names
			if *childService.Id == id {
				return *parentService.Name + " " + *childService.Name, nil
			}
		}
	}

	return "", fmt.Errorf("Service ID %d not found", id)
}

// Details of incident: zone, service, description, start time and type
type EventsDetails [][]string

// Implement Sort interface
func (t EventsDetails) Less(i, j int) bool {
	return t[i][0] < t[j][0]
}
func (t EventsDetails) Len() int {
	return len(t)
}
func (t EventsDetails) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Get the status of a status page
func GetStatusPage(subdomain string) (*StatusPalStatus, error) {
	url := statusPalURL + subdomain + "/summary"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("content-type")
	if contentType != statusContentPage {
		return nil, fmt.Errorf("status page content type expected %q, but got %q", statusContentPage, contentType)
	}

	response := &StatusPalStatus{}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	return response, nil
}
