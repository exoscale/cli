package sks

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/stretchr/testify/require"
)

func TestRemoveAddonFromList(t *testing.T) {
	tests := []struct {
		name          string
		clusterAddons *v3.SKSClusterAddons
		requestAddons *v3.SKSClusterAddons
		addonToRemove string
		expected      v3.SKSClusterAddons
	}{
		{
			name: "Remove Karpenter from list with multiple addons",
			clusterAddons: &v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected: v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name: "Remove addon from list with only that addon",
			clusterAddons: &v3.SKSClusterAddons{
				sksClusterAddonKarpenter,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      v3.SKSClusterAddons{},
		},
		{
			name: "Remove addon that doesn't exist in the list",
			clusterAddons: &v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected: v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name:          "Remove addon from empty cluster addons",
			clusterAddons: &v3.SKSClusterAddons{},
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      v3.SKSClusterAddons{},
		},
		{
			name:          "Remove addon when cluster addons is nil",
			clusterAddons: nil,
			requestAddons: nil,
			addonToRemove: sksClusterAddonKarpenter,
			expected:      v3.SKSClusterAddons{},
		},
		{
			name: "Remove addon when updateReq already has addons",
			clusterAddons: &v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
			},
			requestAddons: &v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonKarpenter,
				sksClusterAddonMetricsServer,
			},
			addonToRemove: sksClusterAddonKarpenter,
			expected: v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonMetricsServer,
			},
		},
		{
			name: "Remove CSI addon from list",
			clusterAddons: &v3.SKSClusterAddons{
				sksClusterAddonExoscaleCCM,
				sksClusterAddonExoscaleCSI,
				sksClusterAddonMetricsServer,
			},
			requestAddons: nil,
			addonToRemove: sksClusterAddonExoscaleCSI,
			expected: v3.SKSClusterAddons{
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
			require.Equal(t, len(tt.expected), len(*updateReq.Addons), "addon list length should match")
			require.ElementsMatch(t, tt.expected, *updateReq.Addons, "addon list should match expected")
		})
	}
}

func TestRemoveAddonFromList_PreservesOrder(t *testing.T) {
	// Test that order is preserved for remaining addons
	clusterAddons := &v3.SKSClusterAddons{
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

	expected := v3.SKSClusterAddons{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
		sksClusterAddonExoscaleCSI,
	}

	require.NotNil(t, updateReq.Addons)
	require.Equal(t, expected, *updateReq.Addons, "order should be preserved")
}

func TestRemoveAddonFromList_MultipleCalls(t *testing.T) {
	// Test that multiple calls work correctly
	clusterAddons := &v3.SKSClusterAddons{
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
	require.Equal(t, 3, len(*updateReq.Addons))

	// Second removal (on already modified list)
	removeAddonFromList(&updateReq, cluster, sksClusterAddonExoscaleCSI)
	require.Equal(t, 2, len(*updateReq.Addons))

	expected := v3.SKSClusterAddons{
		sksClusterAddonExoscaleCCM,
		sksClusterAddonMetricsServer,
	}
	require.ElementsMatch(t, expected, *updateReq.Addons)
}
