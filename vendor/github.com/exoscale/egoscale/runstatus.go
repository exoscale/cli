package egoscale

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// RunstatusErrorResponse represents an error in the API
type RunstatusErrorResponse map[string][]string

// runstatusPagesURL is the only URL that cannot be guessed
const runstatusPagesURL = "/pages"

// RunstatusPage runstatus page
type RunstatusPage struct {
	Created          *time.Time             `json:"created,omitempty"`
	DarkTheme        bool                   `json:"dark_theme,omitempty"`
	Domain           string                 `json:"domain,omitempty"`
	GradientEnd      string                 `json:"gradient_end,omitempty"`
	GradientStart    string                 `json:"gradient_start,omitempty"`
	HeaderBackground string                 `json:"header_background,omitempty"`
	ID               int                    `json:"id,omitempty"`
	Incidents        []RunstatusIncident    `json:"incidents,omitempty"`
	IncidentsURL     string                 `json:"incidents_url,omitempty"`
	Logo             string                 `json:"logo,omitempty"`
	Maintenances     []RunstatusMaintenance `json:"maintenances,omitempty"`
	MaintenancesURL  string                 `json:"maintenances_url,omitempty"`
	Name             string                 `json:"name"` //fake field (used to post a new runstatus page)
	OkText           string                 `json:"ok_text,omitempty"`
	Plan             string                 `json:"plan,omitempty"`
	PublicURL        string                 `json:"public_url,omitempty"`
	Services         []RunstatusService     `json:"services,omitempty"`
	ServicesURL      string                 `json:"services_url,omitempty"`
	State            string                 `json:"state,omitempty"`
	Subdomain        string                 `json:"subdomain"`
	SupportEmail     string                 `json:"support_email,omitempty"`
	TimeZone         string                 `json:"time_zone,omitempty"`
	Title            string                 `json:"title,omitempty"`
	TitleColor       string                 `json:"title_color,omitempty"`
	TwitterUsername  string                 `json:"twitter_username,omitempty"`
	URL              string                 `json:"url,omitempty"`
}

// Match returns true if the other page has got similarities with itself
func (page RunstatusPage) Match(other RunstatusPage) bool {
	if other.Subdomain != "" && page.Subdomain == other.Subdomain {
		return true
	}

	if other.ID > 0 && page.ID == other.ID {
		return true
	}

	return false
}

//RunstatusPageList runstatus page list
type RunstatusPageList struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []RunstatusPage `json:"results"`
}

//RunstatusIncident is a runstatus incident
type RunstatusIncident struct {
	EndDate    *time.Time       `json:"end_date,omitempty"`
	Events     []RunstatusEvent `json:"events,omitempty"`
	EventsURL  string           `json:"events_url,omitempty"`
	ID         int              `json:"id,omitempty"`
	PageURL    string           `json:"page_url,omitempty"` // fake field
	PostMortem string           `json:"post_mortem,omitempty"`
	RealTime   bool             `json:"real_time,omitempty"`
	Services   []string         `json:"services"`
	StartDate  *time.Time       `json:"start_date,omitempty"`
	State      string           `json:"state"`
	Status     string           `json:"status"`
	StatusText string           `json:"status_text"`
	Title      string           `json:"title"`
	URL        string           `json:"url,omitempty"`
}

// Match returns true if the other incident has got similarities with itself
func (incident RunstatusIncident) Match(other RunstatusIncident) bool {
	if other.Title != "" && incident.Title == other.Title {
		return true
	}

	if other.ID > 0 && incident.ID == other.ID {
		return true
	}

	return false
}

//RunstatusIncidentList is a list of incident
type RunstatusIncidentList struct {
	Incidents []RunstatusIncident `json:"incidents"`
}

// RunstatusMaintenance is a runstatus maintenance
type RunstatusMaintenance struct {
	Created     *time.Time       `json:"created,omitempty"`
	Description string           `json:"description,omitempty"`
	EndDate     *time.Time       `json:"end_date"`
	Events      []RunstatusEvent `json:"events,omitempty"`
	EventsURL   string           `json:"events_url,omitempty"`
	ID          int              `json:"id,omitempty"`       // missing field
	PageURL     string           `json:"page_url,omitempty"` // fake field
	RealTime    bool             `json:"real_time,omitempty"`
	Services    []string         `json:"services"`
	StartDate   *time.Time       `json:"start_date"`
	Status      string           `json:"status"`
	Title       string           `json:"title"`
	URL         string           `json:"url,omitempty"`
}

// Match returns true if the other maintenance has got similarities with itself
func (maintenance RunstatusMaintenance) Match(other RunstatusMaintenance) bool {
	if other.Title != "" && maintenance.Title == other.Title {
		return true
	}

	if other.ID > 0 && maintenance.ID == other.ID {
		return true
	}

	return false
}

// FakeID fills up the ID field as it's currently missing
func (maintenance *RunstatusMaintenance) FakeID() error {
	if maintenance.URL == "" {
		return fmt.Errorf("empty URL for %#v", maintenance)
	}

	u, err := url.Parse(maintenance.URL)
	if err != nil {
		return err
	}

	s := path.Base(u.Path)
	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	maintenance.ID = id
	return nil
}

//RunstatusMaintenanceList is a list of incident
type RunstatusMaintenanceList struct {
	Maintenances []RunstatusMaintenance `json:"maintenances"`
}

//RunstatusEvent is a runstatus event
type RunstatusEvent struct {
	Created *time.Time `json:"created,omitempty"`
	State   string     `json:"state,omitempty"`
	Status  string     `json:"status"`
	Text    string     `json:"text"`
}

// RunstatusService is a runstatus service
type RunstatusService struct {
	ID      int    `json:"id"` // missing field
	Name    string `json:"name"`
	PageURL string `json:"page_url,omitempty"` // fake field
	State   string `json:"state,omitempty"`
	URL     string `json:"url,omitempty"`
}

// FakeID fills up the ID field as it's currently missing
func (service *RunstatusService) FakeID() error {
	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	u, err := url.Parse(service.URL)
	if err != nil {
		return err
	}

	s := path.Base(u.Path)
	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	service.ID = id
	return nil
}

// Match returns true if the other service has got similarities with itself
func (service RunstatusService) Match(other RunstatusService) bool {
	if other.Name != "" && service.Name == other.Name {
		return true
	}

	if other.ID > 0 && service.ID == other.ID {
		return true
	}

	return false
}

// RunstatusServiceList service list
type RunstatusServiceList struct {
	Services []RunstatusService `json:"services"`
}

// DeleteRunstatusService delete runstatus service
func (client *Client) DeleteRunstatusService(ctx context.Context, service RunstatusService) error {
	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	_, err := client.runstatusRequest(ctx, service.URL, nil, "DELETE")
	return err
}

// CreateRunstatusService create runstatus service
func (client *Client) CreateRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", service)
	}

	page, err := client.GetRunstatusPage(ctx, RunstatusPage{URL: service.PageURL})
	if err != nil {
		return nil, err
	}

	resp, err := client.runstatusRequest(ctx, page.ServicesURL, service, "POST")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, s); err != nil {
		return nil, err
	}

	return s, nil
}

// GetRunstatusService displays service detail.
func (client *Client) GetRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.URL != "" {
		return client.getRunstatusService(ctx, service)
	}

	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL in %#v", service)
	}

	page, err := client.GetRunstatusPage(ctx, RunstatusPage{URL: service.PageURL})
	if err != nil {
		return nil, err
	}

	ss, err := client.ListRunstatusServices(ctx, *page)
	if err != nil {
		return nil, err
	}

	for i := range ss {
		if ss[i].Match(service) {
			return &ss[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", service)
}

func (client *Client) getRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.URL == "" {
		return nil, fmt.Errorf("missing URL for %#v", service)
	}

	resp, err := client.runstatusRequest(ctx, service.URL, nil, "GET")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, &s); err != nil {
		return nil, err
	}

	if err := s.FakeID(); err != nil {
		log.Printf("bad fake ID for %#v, %s", s, err)
	}

	return s, nil
}

// ListRunstatusServices displays the list of services.
func (client *Client) ListRunstatusServices(ctx context.Context, page RunstatusPage) ([]RunstatusService, error) {
	if page.ServicesURL == "" {
		return nil, fmt.Errorf("empty Services URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.ServicesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusServiceList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Services {
		if err := p.Services[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Services[i], err)
		}
	}

	// NOTE: no pagination
	return p.Services, nil
}

// CreateRunstatusEvent create runstatus incident event
func (client *Client) CreateRunstatusEvent(ctx context.Context, incident RunstatusIncident, event RunstatusEvent) error {
	if incident.EventsURL == "" {
		return fmt.Errorf("empty Events URL for %#v", incident)
	}

	_, err := client.runstatusRequest(ctx, incident.EventsURL, event, "POST")
	return err
}

// GetRunstatusMaintenance retrieves the details of a specific maintenance.
func (client *Client) GetRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) (*RunstatusMaintenance, error) {
	if maintenance.URL != "" {
		return client.getRunstatusMaintenance(ctx, maintenance)
	}

	if maintenance.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", maintenance)
	}

	page, err := client.getRunstatusPage(ctx, RunstatusPage{URL: maintenance.PageURL})
	if err != nil {
		return nil, err
	}

	ms, err := client.ListRunstatusMaintenances(ctx, *page)
	if err != nil {
		return nil, err
	}

	for i := range ms {
		if ms[i].Match(maintenance) {
			return &ms[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", maintenance)
}

func (client *Client) getRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) (*RunstatusMaintenance, error) {
	if maintenance.URL == "" {
		return nil, fmt.Errorf("empty URL for %#v", maintenance)
	}

	resp, err := client.runstatusRequest(ctx, maintenance.URL, nil, "GET")
	if err != nil {
		return nil, err
	}

	m := new(RunstatusMaintenance)
	if err := json.Unmarshal(resp, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListRunstatusMaintenances returns the list of maintenances for the page.
func (client *Client) ListRunstatusMaintenances(ctx context.Context, page RunstatusPage) ([]RunstatusMaintenance, error) {
	if page.MaintenancesURL == "" {
		return nil, fmt.Errorf("empty Maintenances URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.MaintenancesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusMaintenanceList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Maintenances {
		if err := p.Maintenances[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Maintenances[i], err)
		}
	}

	// NOTE: the list of maintenances doesn't have any pagination
	return p.Maintenances, nil
}

// CreateRunstatusMaintenance create runstatus Maintenance
func (client *Client) CreateRunstatusMaintenance(ctx context.Context, page RunstatusPage, maintenance RunstatusMaintenance) error {
	if page.MaintenancesURL == "" {
		return fmt.Errorf("empty Maintenances URL for %#v", page)
	}

	_, err := client.runstatusRequest(ctx, page.MaintenancesURL, maintenance, "POST")
	return err
}

// DeleteRunstatusMaintenance delete runstatus Maintenance
func (client *Client) DeleteRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) error {
	if maintenance.URL == "" {
		return fmt.Errorf("empty URL for %#v", maintenance)
	}

	_, err := client.runstatusRequest(ctx, maintenance.URL, nil, "DELETE")
	return err
}

// UpdateRunstatusMaintenance adds a event to a maintenance.
// Events can be updates or final message with status completed.
func (client *Client) UpdateRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance, event RunstatusEvent) error {
	if maintenance.EventsURL == "" {
		return fmt.Errorf("empty Events URL for %#v", maintenance)
	}

	_, err := client.runstatusRequest(ctx, maintenance.EventsURL, event, "POST")
	return err
}

// GetRunstatusIncident retrieves the details of a specific incident.
func (client *Client) GetRunstatusIncident(ctx context.Context, incident RunstatusIncident) (*RunstatusIncident, error) {
	if incident.URL != "" {
		return client.getRunstatusIncident(ctx, incident)
	}

	if incident.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", incident)
	}

	page, err := client.getRunstatusPage(ctx, RunstatusPage{URL: incident.PageURL})
	if err != nil {
		return nil, err
	}

	is, err := client.ListRunstatusIncidents(ctx, *page)
	if err != nil {
		return nil, err
	}

	for i := range is {
		if is[i].Match(incident) {
			return &is[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", incident)
}

func (client *Client) getRunstatusIncident(ctx context.Context, incident RunstatusIncident) (*RunstatusIncident, error) {
	if incident.URL == "" {
		return nil, fmt.Errorf("empty URL for %#v", incident)
	}

	resp, err := client.runstatusRequest(ctx, incident.URL, nil, "GET")
	if err != nil {
		return nil, err
	}

	i := new(RunstatusIncident)
	if err := json.Unmarshal(resp, i); err != nil {
		return nil, err
	}
	return i, nil
}

// ListRunstatusIncidents lists the incidents for a specific page.
func (client *Client) ListRunstatusIncidents(ctx context.Context, page RunstatusPage) ([]RunstatusIncident, error) {
	if page.IncidentsURL == "" {
		return nil, fmt.Errorf("empty Incidents URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.IncidentsURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusIncidentList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// NOTE: no pagination
	return p.Incidents, nil
}

// CreateRunstatusIncident create runstatus incident
func (client *Client) CreateRunstatusIncident(ctx context.Context, page RunstatusPage, incident RunstatusIncident) error {
	if page.IncidentsURL == "" {
		return fmt.Errorf("empty Incidents URL for %#v", page)
	}

	_, err := client.runstatusRequest(ctx, page.IncidentsURL, incident, "POST")
	return err
}

// DeleteRunstatusIncident delete runstatus incident
func (client *Client) DeleteRunstatusIncident(ctx context.Context, incident RunstatusIncident) error {
	if incident.URL == "" {
		return fmt.Errorf("empty URL for %#v", incident)
	}

	_, err := client.runstatusRequest(ctx, incident.URL, nil, "DELETE")
	return err
}

// CreateRunstatusPage create runstatus page
func (client *Client) CreateRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, page, "POST")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPage
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteRunstatusPage delete runstatus page
func (client *Client) DeleteRunstatusPage(ctx context.Context, page RunstatusPage) error {
	if page.URL != "" {
		return fmt.Errorf("empty URL for %#v", page)
	}
	_, err := client.runstatusRequest(ctx, page.URL, nil, "DELETE")
	return err
}

// GetRunstatusPage fetches the runstatus page
func (client *Client) GetRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	if page.URL != "" {
		return client.getRunstatusPage(ctx, page)
	}

	ps, err := client.ListRunstatusPages(ctx)
	if err != nil {
		return nil, err
	}

	for i := range ps {
		if ps[i].Match(page) {
			return &ps[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", page)
}

func (client *Client) getRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	if page.URL == "" {
		return nil, fmt.Errorf("empty URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.URL, nil, "GET")
	if err != nil {
		return nil, err
	}

	p := new(RunstatusPage)
	if err := json.Unmarshal(resp, p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Maintenances {
		if err := p.Maintenances[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Maintenances[i], err)
		}
	}

	return p, nil
}

// ListRunstatusPages list all the runstatus pages
func (client *Client) ListRunstatusPages(ctx context.Context) ([]RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPageList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// XXX: handle pagination
	return p.Results, nil
}

// Error formats the DNSerror into a string
func (req RunstatusErrorResponse) Error() string {
	if len(req) > 0 {
		errs := []string{}
		for name, ss := range req {
			if len(ss) > 0 {
				errs = append(errs, fmt.Sprintf("%s: %s", name, strings.Join(ss, ", ")))
			}
		}
		return fmt.Sprintf("Runstatus error: %s", strings.Join(errs, "; "))
	}
	return fmt.Sprintf("Runstatus error")
}

func (client *Client) runstatusRequest(ctx context.Context, uri string, structParam interface{}, method string) (json.RawMessage, error) {
	reqURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if reqURL.Scheme == "" {
		return nil, fmt.Errorf("only absolute URI are considered valid, got %q", uri)
	}

	var params string
	if structParam != nil {
		m, err := json.Marshal(structParam)
		if err != nil {
			return nil, err
		}
		params = string(m)
	}

	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	time := time.Now().Local().Format("2006-01-02T15:04:05-0700")

	payload := fmt.Sprintf("%s%s%s", req.URL.String(), time, params)

	mac := hmac.New(sha256.New, []byte(client.apiSecret))
	_, err = mac.Write([]byte(payload))
	if err != nil {
		return nil, err
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	var hdr = make(http.Header)

	hdr.Add("Authorization", fmt.Sprintf("Exoscale-HMAC-SHA256 %s:%s", client.APIKey, signature))
	hdr.Add("Exoscale-Date", time)
	hdr.Add("User-Agent", fmt.Sprintf("exoscale/egoscale (%v)", Version))
	hdr.Add("Accept", "application/json")
	if params != "" {
		hdr.Add("Content-Type", "application/json")
	}
	req.Header = hdr

	req = req.WithContext(ctx)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	if resp.StatusCode == 204 {
		if method != "DELETE" {
			return nil, fmt.Errorf("only DELETE is expected to produce 204, was %q", method)
		}
		return nil, nil
	}

	contentType := resp.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf(`response content-type expected to be "application/json", got %q`, contentType)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("not found")
		}
		e := new(RunstatusErrorResponse)
		if err := json.Unmarshal(b, e); err != nil {
			return nil, err
		}
		return nil, e
	}

	return b, nil
}
