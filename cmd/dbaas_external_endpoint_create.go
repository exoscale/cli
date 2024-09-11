package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type dbaasExternalEndpointCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Type string `cli-arg:"#"`
	Name string `cli-arg:"#"`

	HelpDatadog       bool `cli-usage:"show usage for flags specific to the datadog external endpoint type"`
	HelpElasticsearch bool `cli-usage:"show usage for flags specific to the elasticsearch external endpoint type"`
	HelpOpensearch    bool `cli-usage:"show usage for flags specific to the opensearch external endpoint type"`
	HelpPrometheus    bool `cli-usage:"show usage for flags specific to the prometheus external endpoint type"`
	HelpRsyslog       bool `cli-usage:"show usage for flags specific to the rsyslog external endpoint type"`

	DatadogAPIKey                      string `cli-flag:"datadog-api-key" cli-usage:"Datadog API key" cli-hidden:""`
	DatadogSite                        string `cli-flag:"datadog-site" cli-usage:"Datadog intake site. Defaults to datadoghq.com" cli-hidden:""`
	DatadogTags                        string `cli-flag:"datadog-tags" cli-usage:"Datadog tags. Example. '[{\"comment\": \"ex\", \"tag\": \"aiven-asdfasda\"}]'" cli-hidden:""`
	DatadogDisableConsumerStats        bool   `cli-flag:"datadog-disable-consumer-stats" cli-usage:"Disable consumer group metrics" cli-hidden:""`
	DatadogKafkaConsumerCheckInstances int64  `cli-flag:"datadog-kafka-consumer-check-instances" cli-usage:"Number of separate instances to fetch kafka consumer statistics with" cli-hidden:""`
	DatadogKafkaConsumerStatsTimeout   int64  `cli-flag:"datadog-kafka-consumer-stats-timeout" cli-usage:"Number of seconds that datadog will wait to get consumer statistics from brokers" cli-hidden:""`
	DatadogMaxPartitionContexts        int64  `cli-flag:"datadog-max-partition-contexts" cli-usage:"Maximum number of partition contexts to send" cli-hidden:""`

	ElasticsearchURL          string `cli-flag:"elasticsearch-url" cli-usage:"Elasticsearch connection URL" cli-hidden:""`
	ElasticsearchIndexPrefix  string `cli-flag:"elasticsearch-index-prefix" cli-usage:"Elasticsearch index prefix" cli-hidden:""`
	ElasticsearchCA           string `cli-flag:"elasticsearch-ca" cli-usage:"PEM encoded CA certificate" cli-hidden:""`
	ElasticsearchIndexDaysMax int64  `cli-flag:"elasticsearch-index-days-max" cli-usage:"Maximum number of days of logs to keep" cli-hidden:""`
	ElasticsearchTimeout      int64  `cli-flag:"elasticsearch-timeout" cli-usage:"Elasticsearch request timeout limit" cli-hidden:""`

	OpensearchURL          string `cli-flag:"opensearch-url" cli-usage:"OpenSearch connection URL" cli-hidden:""`
	OpensearchIndexPrefix  string `cli-flag:"opensearch-index-prefix" cli-usage:"OpenSearch index prefix" cli-hidden:""`
	OpensearchCA           string `cli-flag:"opensearch-ca" cli-usage:"PEM encoded CA certificate" cli-hidden:""`
	OpensearchIndexDaysMax int64  `cli-flag:"opensearch-index-days-max" cli-usage:"Maximum number of days of logs to keep" cli-hidden:""`
	OpensearchTimeout      int64  `cli-flag:"opensearch-timeout" cli-usage:"OpenSearch request timeout limit" cli-hidden:""`

	PrometheusBasicAuthUsername string `cli-flag:"prometheus-basic-auth-username" cli-usage:"Prometheus basic authentication username" cli-hidden:""`
	PrometheusBasicAuthPassword string `cli-flag:"prometheus-basic-auth-password" cli-usage:"Prometheus basic authentication password" cli-hidden:""`

	RsyslogServer         string `cli-flag:"rsyslog-server" cli-usage:"Rsyslog server IP address or hostname" cli-hidden:""`
	RsyslogPort           int64  `cli-flag:"rsyslog-port" cli-usage:"Rsyslog server port" cli-hidden:""`
	RsyslogTls            bool   `cli-flag:"rsyslog-tls" cli-usage:"Require TLS" cli-hidden:""`
	RsyslogFormat         string `cli-flag:"rsyslog-format" cli-usage:"Message format" cli-hidden:""`
	RsyslogKey            string `cli-flag:"rsyslog-key" cli-usage:"PEM encoded client key" cli-hidden:""`
	RsyslogLogline        string `cli-flag:"rsyslog-logline" cli-usage:"Custom syslog message format" cli-hidden:""`
	RsyslogCA             string `cli-flag:"rsyslog-ca" cli-usage:"PEM encoded CA certificate" cli-hidden:""`
	RsyslogCert           string `cli-flag:"rsyslog-cert" cli-usage:"PEM encoded client certificate" cli-hidden:""`
	RsyslogSD             string `cli-flag:"rsyslog-sd" cli-usage:"Structured data block for log message" cli-hidden:""`
	RsyslogMaxMessageSize int64  `cli-flag:"rsyslog-max-message-size" cli-usage:"Rsyslog max message size" cli-hidden:""`
}

func (c *dbaasExternalEndpointCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	switch {
	case cmd.Flags().Changed("help-datadog"):
		cmdShowHelpFlags(cmd.Flags(), "datadog-")
		os.Exit(0)
	case cmd.Flags().Changed("help-elasticsearch"):
		cmdShowHelpFlags(cmd.Flags(), "elasticsearch-")
		os.Exit(0)
	case cmd.Flags().Changed("help-opensearch"):
		cmdShowHelpFlags(cmd.Flags(), "opensearch-")
		os.Exit(0)
	case cmd.Flags().Changed("help-prometheus"):
		cmdShowHelpFlags(cmd.Flags(), "prometheus-")
		os.Exit(0)
	case cmd.Flags().Changed("help-rsyslog"):
		cmdShowHelpFlags(cmd.Flags(), "rsyslog-")
		os.Exit(0)
	}

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasExternalEndpointCreateCmd) cmdAliases() []string {
	return gCreateAlias
}

func (c *dbaasExternalEndpointCreateCmd) cmdLong() string {
	return "Create a new external endpoint for DBaaS"
}

func (c *dbaasExternalEndpointCreateCmd) cmdShort() string {
	return "Create a new external endpoint for DBaaS"
}

func (c *dbaasExternalEndpointCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	// Implement the command's main logic here
	switch c.Type {
	case "datadog":
		return c.createDatadog(cmd, args)
	// case "elasticsearch":
	// 	return c.createElasticsearch(cmd, args)
	case "opensearch":
		return c.createOpensearch(cmd, args)
	// case "prometheus":
	// 	return c.createPrometheus(cmd, args)
	// case "rsyslog":
	// 	return c.createRsyslog(cmd, args)
	default:
		return fmt.Errorf("unsupported external endpoint type %q", c.Type)
	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasExternalEndpointCmd, &dbaasExternalEndpointCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
