package dbaas

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

// TestReadReplicaGuardLogic tests that the read replica detection logic
// correctly identifies services with active read replicas that should block deletion.
func TestReadReplicaGuardLogic(t *testing.T) {
	tests := []struct {
		name                 string
		service              v3.DBAASServiceCommon
		expectBlock          bool
		expectedReplicaNames []string
	}{
		{
			name: "no integrations - should allow delete",
			service: v3.DBAASServiceCommon{
				Name:         "test-service",
				Integrations: []v3.DBAASIntegration{},
			},
			expectBlock: false,
		},
		{
			name: "nil integrations - should allow delete",
			service: v3.DBAASServiceCommon{
				Name:         "test-service",
				Integrations: nil,
			},
			expectBlock: false,
		},
		{
			name: "non-read-replica integration - should allow delete",
			service: v3.DBAASServiceCommon{
				Name: "test-service",
				Integrations: []v3.DBAASIntegration{
					{
						Type:     "prometheus",
						ISActive: ptrBool(true),
						Source:   "test-service",
						Dest:     "monitoring",
					},
				},
			},
			expectBlock: false,
		},
		{
			name: "read replica inactive - should allow delete",
			service: v3.DBAASServiceCommon{
				Name: "test-service",
				Integrations: []v3.DBAASIntegration{
					{
						Type:     "read_replica",
						ISActive: ptrBool(false),
						Source:   "test-service",
						Dest:     "replica-1",
					},
				},
			},
			expectBlock: false,
		},
		{
			name: "read replica active - should block delete",
			service: v3.DBAASServiceCommon{
				Name: "test-service",
				Integrations: []v3.DBAASIntegration{
					{
						Type:     "read_replica",
						ISActive: ptrBool(true),
						Source:   "test-service",
						Dest:     "replica-1",
					},
				},
			},
			expectBlock:          true,
			expectedReplicaNames: []string{"replica-1"},
		},
		{
			name: "multiple read replicas mixed - should block with active ones",
			service: v3.DBAASServiceCommon{
				Name: "test-service",
				Integrations: []v3.DBAASIntegration{
					{
						Type:     "read_replica",
						ISActive: ptrBool(true),
						Source:   "test-service",
						Dest:     "replica-1",
					},
					{
						Type:     "read_replica",
						ISActive: ptrBool(false),
						Source:   "test-service",
						Dest:     "replica-2",
					},
					{
						Type:     "read_replica",
						ISActive: ptrBool(true),
						Source:   "test-service",
						Dest:     "replica-3",
					},
				},
			},
			expectBlock:          true,
			expectedReplicaNames: []string{"replica-1", "replica-3"},
		},
		{
			name: "read replica with nil ISActive - should allow (defaults to false)",
			service: v3.DBAASServiceCommon{
				Name: "test-service",
				Integrations: []v3.DBAASIntegration{
					{
						Type:     "read_replica",
						ISActive: nil,
						Source:   "test-service",
						Dest:     "replica-1",
					},
				},
			},
			expectBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readReplicaNames := dbaasActiveReadReplicaNamesForPrimary(tt.service)
			shouldBlock := len(readReplicaNames) > 0

			if shouldBlock != tt.expectBlock {
				t.Errorf("expected block=%v, got block=%v (replicas=%v)", tt.expectBlock, shouldBlock, readReplicaNames)
			}

			if tt.expectBlock && len(tt.expectedReplicaNames) > 0 {
				if len(readReplicaNames) != len(tt.expectedReplicaNames) {
					t.Errorf("expected replica names %v, got %v", tt.expectedReplicaNames, readReplicaNames)
				}
			}
		})
	}
}

func ptrBool(b bool) *bool {
	return &b
}
