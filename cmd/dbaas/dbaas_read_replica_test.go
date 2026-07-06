package dbaas

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

func TestGetReadReplicaIntegrationForReplica(t *testing.T) {
	replicaService := v3.DBAASServiceCommon{
		Name: "replica-1",
		Integrations: []v3.DBAASIntegration{
			{
				Type:   "read_replica",
				Source: "primary-1",
				Dest:   "replica-1",
			},
		},
	}

	replicaIntegration := dbaasGetReadReplicaIntegrationForReplica(replicaService)
	if replicaIntegration == nil {
		t.Fatal("expected to find read replica integration for replica service")
	}
	if replicaIntegration.Source != "primary-1" {
		t.Fatalf("expected source service primary-1, got %q", replicaIntegration.Source)
	}

	primaryService := v3.DBAASServiceCommon{
		Name: "primary-1",
		Integrations: []v3.DBAASIntegration{
			{
				Type:   "read_replica",
				Source: "primary-1",
				Dest:   "replica-1",
			},
		},
	}

	if dbaasGetReadReplicaIntegrationForReplica(primaryService) != nil {
		t.Fatal("expected primary service not to be detected as a replica")
	}
}

func TestReadReplicaSupportedServiceType(t *testing.T) {
	tests := []struct {
		serviceType string
		expected    bool
	}{
		{serviceType: "pg", expected: true},
		{serviceType: "mysql", expected: true},
		{serviceType: "kafka", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.serviceType, func(t *testing.T) {
			if actual := dbaasReadReplicaSupportedServiceType(tt.serviceType); actual != tt.expected {
				t.Fatalf("expected %v for %q, got %v", tt.expected, tt.serviceType, actual)
			}
		})
	}
}
