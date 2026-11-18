package internal

// Settings groups supported by each Database Service type.
var (
	GrafanaSettings    = []string{"grafana"}
	OpensearchSettings = []string{"opensearch"}
	KafkaSettings      = []string{
		"kafka",
		"kafka-rest",
		"kafka-connect",
		"schema-registry",
	}
	MysqlSettings = []string{"mysql"}
	PGSettings    = []string{
		"pg",
		"pgbouncer",
		"pglookout",
		"timescaledb",
	}
	ValkeySettings     = []string{"valkey"}
	ThanosSettings     = []string{"thanos", "compactor", "query", "query-frontend"}
	ClickhouseSettings = []string{"clickhouse"}
)
