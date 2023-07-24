package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// https://www.statuspal.io/api-docs
const (
	statusPalURL      = "https://statuspal.eu/api/v1/status_pages/"
	statusContentPage = "application/json; charset=utf-8"
	dateLayout        = "2006-01-02T15:04:05"
)

// https://www.statuspal.io/api-docs#tag/Status/operation/getStatusPageStatus
type StatusPalStatus struct {
	// Services: all the (parent) services of the StatusPage with the current incident type
	Services []Service `json:"services"`

	// Maintenances only contains the future scheduled maintenances
	// Ongoing maintenances are in Incidents
	// Incidents
	Incidents []Incident `json:"incidents"`
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
	IncidentType *IncidentType `json:"current_incident_type,omitempty"`

	// Each product available in the zone
	Children *[]Service `json:"children,omitempty"`
}

type Incident struct {
	Id    *int    `json:"id,omitempty"`
	Title *string `json:"title,omitempty"`
	// The time at which the incident/maintenance started(UTC).
	//StartsAt time.Time `json:"starts_at"`
	StartsAt string `json:"starts_at"`
	// Type of current incident (major, minor, scheduled)
	Type IncidentType `json:"type"`
	// Services impacted (only id and name)
	Services []Service `json:"services"`
}

const (
	IncidentTypeMajor     IncidentType = "major"
	IncidentTypeMinor     IncidentType = "minor"
	IncidentTypeScheduled IncidentType = "scheduled"
)

type Services []Service

type IncidentType string

// Details of incident: zone, service, description, start time and type
type IncidentsDetails [][]string

// Implement Sort interface
func (t IncidentsDetails) Less(i, j int) bool {
	return t[i][0] < t[j][0]
}

func (t IncidentsDetails) Len() int {
	return len(t)
}

func (t IncidentsDetails) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Get the status of a status page
func GetStatusPage(subdomain string) (*StatusPalStatus, error) {
	url := statusPalURL + subdomain + "/status"
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

// Get the status of global services (status of global or zones, no details of impacted services)
func (s StatusPalStatus) GetStatusByZone() ([][]string, error) {
	var global [][]string
	for _, svc := range s.Services {
		state := svc.getIncidentType()
		global = append(global, []string{*svc.Name, state})
	}
	return global, nil
}

func (s Services) isParentService(id int) bool {
	for _, service := range s {
		if *service.Id == id {
			return true
		}
	}
	return false

}

// Return the Zone and the impacted service = fullname (parent svc + child svc)
func (s Services) getServiceNamebyId(id int) (string, error) {
	// For all zones / global services
	for _, parentService := range s {
		// id provided is a parent service, return the name
		if *parentService.Id == id {
			return *parentService.Name, nil
		}
		// Try to find the Service Id in the child services
		for _, childService := range *parentService.Children {
			// In this case, we returen the Parent and the Child names
			if *childService.Id == id {
				return *parentService.Name + " " + *childService.Name, nil
			}
		}
	}

	return "", fmt.Errorf("Service ID %d not found", id)
}

func (s StatusPalStatus) GetIncidents() ([][]string, [][]string, error) {
	var incidents IncidentsDetails
	var maintenances IncidentsDetails
	var services Services
	services = s.Services
	// In Incidents, we have maintenances and incidents currently taking place
	// We need to show them in different tables
	for _, event := range s.Incidents {
		// Get all the services impacted by the incident (name and id)
		// Child and parent are all mixed, we need to rebuild the dependency
		for _, impacted := range event.Services {
			if services.isParentService(*impacted.Id) {
				continue
			}
			svcName, err := services.getServiceNamebyId(*impacted.Id)
			if err != nil {
				return nil, nil, err
			}
			started, err := time.Parse(dateLayout, *&event.StartsAt)
			if err != nil {
				return nil, nil, err
			}
			startTimeUTC := started.Format(time.RFC822)
			eventDetails := []string{svcName, *event.Title}
			if event.Type == IncidentTypeScheduled {
				eventDetails = append(eventDetails, "scheduled at "+startTimeUTC)
				maintenances = append(maintenances, eventDetails)
			} else {

				eventDetails = append(eventDetails, fmt.Sprint(event.Type), "since "+startTimeUTC)
				incidents = append(incidents, eventDetails)
			}
		}
	}
	// Sort by zones
	sort.Sort(incidents)
	sort.Sort(maintenances)
	return incidents, maintenances, nil
}
