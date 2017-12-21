/*

APIs

All the available APIs on the server and provided by the API Discovery plugin

	cs := egoscale.NewClient("https://api.exoscale.ch/compute", "EXO...", "...")

	resp := new(egoscale.ListApisResponse)
	err := cs.Request(&egoscale.ListApisRequest{}, resp)
	if err != nil {
		panic(err)
	}

	for _, api := range resp.Api {
		fmt.Println("%s %s", api.Name, api.Description)
	}
	// Output:
	// listNetworks Lists all available networks
	// ...

*/

package egoscale

// Api represents an API service
type Api struct {
	Description string         `json:"description"`
	IsAsync     bool           `json:"isasync"`
	Name        string         `json:"name"`
	Related     string         `json:"related"` // comma separated
	Since       string         `json:"since"`
	Type        string         `json:"type"`
	Params      []*ApiParam    `json:"params"`
	Response    []*ApiResponse `json:"responses"`
}

// ApiParam represents an API parameter field
type ApiParam struct {
	Description string `json:"description"`
	Length      int64  `json:"length"`
	Name        string `json:"name"`
	Related     string `json:"related"` // comma separated
	Since       string `json:"since"`
	Type        string `json:"type"`
}

// ApiResponse represents an API response field
type ApiResponse struct {
	Description string         `json:"description"`
	Name        string         `json:"name"`
	Response    []*ApiResponse `json:"response"`
	Type        string         `json:"type"`
}

// ListApisRequest represents a query to list the api
type ListApisRequest struct {
	Name string `json:"name,omitempty"`
}

// Command returns the CloudStack API command
func (req *ListApisRequest) Command() string {
	return "listApis"
}

// ListApisResponse represents a list of API
type ListApisResponse struct {
	Count int    `json:"count"`
	Api   []*Api `json:"api"`
}
