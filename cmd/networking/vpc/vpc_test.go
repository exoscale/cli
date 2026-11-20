package vpc

import (
	"strings"
	"testing"
)

func TestValidateAttachFlags(t *testing.T) {
	tests := []struct {
		name        string
		vpc         string
		subnetCount int
		wantErr     string
	}{
		{name: "no vpc flags"},
		{name: "vpc and subnet", vpc: "prod", subnetCount: 1},
		{name: "vpc and subnets", vpc: "prod", subnetCount: 2},
		{name: "subnet without vpc", subnetCount: 1, wantErr: "--vpc is required when --vpc-subnet is specified"},
		{name: "vpc without subnet", vpc: "prod", wantErr: "at least one --vpc-subnet is required when --vpc is specified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAttachFlags(tt.vpc, tt.subnetCount)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
