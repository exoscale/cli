package dbaasutils

// External endpoint types, shared by the external-endpoint and
// external-integration commands.
const (
	ExternalEndpointTypeDatadog       = "datadog"
	ExternalEndpointTypeOpensearch    = "opensearch"
	ExternalEndpointTypeElasticsearch = "elasticsearch"
	ExternalEndpointTypePrometheus    = "prometheus"
	ExternalEndpointTypeRsyslog       = "rsyslog"
)
