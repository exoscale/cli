package sks

import (
	"encoding/json"
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/stretchr/testify/require"
)

func TestRemoveAddonFromList(t *testing.T) {
	tests := []struct {
		name          string
		clusterAddons []string
		requestAddons []string
		addonToRemove string
		expected      []string
	}{
		{
			name: "Remove Karpenter from list with multiple addons",
			clusterAddons: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name: "Remove addon from list with only that addon",
			clusterAddons: []string{
				sksClusterAddonKarpenter,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      []string{},
		},
		{
			name: "Remove addon that doesn't exist in the list",
			clusterAddons: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name:          "Remove addon from empty cluster addons",
			clusterAddons: []string{},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      []string{},
		},
		{
			name:          "Remove addon when cluster addons is nil",
			clusterAddons: nil,
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      []string{},
		},
		{
			name: "Remove addon when updateReq already has addons",
			clusterAddons: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
			},
			requestAddons: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
				sksClusterAddonMetricsServer,
			},
			addonToRemove: sksClusterAddonKarpenter,
			expected: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name: "Remove CSI addon from list",
			clusterAddons: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonExoscaleCSI,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonExoscaleCSI,
			expected: []string{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cluster := v3.SKSCluster{
				Addons: tt.clusterAddons,
			}
			updateReq := v3.UpdateSKSClusterRequest{
				Addons: tt.requestAddons,
			}

			// Execute
			removeAddonFromList(&updateReq, cluster, tt.addonToRemove)

			// Assert
			require.NotNil(t, updateReq.Addons, "updateReq.Addons should not be nil")
			require.Equal(t, len(tt.expected), len(updateReq.Addons), "addon list length should match")
			require.ElementsMatch(t, tt.expected, updateReq.Addons, "addon list should match expected")
		})
	}
}

func TestRemoveAddonFromList_PreservesOrder(t *testing.T) {
	// Test that order is preserved for remaining addons
	clusterAddons := []string{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
		sksClusterAddonKarpenter,
		sksClusterAddonExoscaleCSI,
	}

	cluster := v3.SKSCluster{
		Addons: clusterAddons,
	}
	updateReq := v3.UpdateSKSClusterRequest{}

	removeAddonFromList(&updateReq, cluster, sksClusterAddonKarpenter)

	expected := []string{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
		sksClusterAddonExoscaleCSI,
	}

	require.NotNil(t, updateReq.Addons)
	require.Equal(t, expected, updateReq.Addons, "order should be preserved")
}

func TestRemoveAddonFromList_MultipleCalls(t *testing.T) {
	// Test that multiple calls work correctly
	clusterAddons := []string{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
		sksClusterAddonKarpenter,
		sksClusterAddonExoscaleCSI,
	}

	cluster := v3.SKSCluster{
		Addons: clusterAddons,
	}
	updateReq := v3.UpdateSKSClusterRequest{}

	// First removal
	removeAddonFromList(&updateReq, cluster, sksClusterAddonKarpenter)
	require.Equal(t, 3, len(updateReq.Addons))

	// Second removal (on already modified list)
	removeAddonFromList(&updateReq, cluster, sksClusterAddonExoscaleCSI)
	require.Equal(t, 2, len(updateReq.Addons))

	expected := []string{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
	}
	require.ElementsMatch(t, expected, updateReq.Addons)
}

// TestKarpenterFeatureGatesRoundTrip ensures the KarpenterFeatureGates field on the
// SKSCluster and UpdateSKSClusterRequest wire types is properly serialized
// using the `karpenter-feature-gates` JSON tag, matching FeatureGates placement.
func TestKarpenterFeatureGatesRoundTrip(t *testing.T) {
	t.Parallel()

	cluster := v3.SKSCluster{
		KarpenterFeatureGates: []string{"Drift=true", "SpotToSpotConsolidation=false"},
	}
	data, err := json.Marshal(cluster)
	require.NoError(t, err)
	require.Contains(t, string(data), `"karpenter-feature-gates":["Drift=true","SpotToSpotConsolidation=false"]`)

	updateReq := v3.UpdateSKSClusterRequest{
		KarpenterFeatureGates: []string{"Drift=true"},
	}
	data, err = json.Marshal(updateReq)
	require.NoError(t, err)
	require.Contains(t, string(data), `"karpenter-feature-gates":["Drift=true"]`)

	createReq := v3.CreateSKSClusterRequest{
		Name:                  "test",
		Level:                 v3.CreateSKSClusterRequestLevelPro,
		Version:               "1.30.0",
		KarpenterFeatureGates: []string{"Drift=true"},
	}
	data, err = json.Marshal(createReq)
	require.NoError(t, err)
	require.Contains(t, string(data), `"karpenter-feature-gates":["Drift=true"]`)
}
