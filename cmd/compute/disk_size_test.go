package compute

import "testing"

func TestResolveTemplateDiskSize(t *testing.T) {
	tests := []struct {
		name          string
		diskSize      int64
		templateSize  int64
		explicitlySet bool
		want          int64
		wantErr       string
	}{
		{
			name:         "default is large enough",
			diskSize:     50,
			templateSize: 10 * gibibyte,
			want:         50,
		},
		{
			name:         "default is increased",
			diskSize:     50,
			templateSize: 80 * gibibyte,
			want:         80,
		},
		{
			name:         "template size is rounded up",
			diskSize:     50,
			templateSize: 80*gibibyte + 1,
			want:         81,
		},
		{
			name:          "explicit size is rejected",
			diskSize:      50,
			templateSize:  80 * gibibyte,
			explicitlySet: true,
			wantErr:       "--disk-size 50 GiB is smaller than the selected template's minimum of 80 GiB; use --disk-size 80 or larger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTemplateDiskSize(tt.diskSize, tt.templateSize, tt.explicitlySet)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("ResolveTemplateDiskSize() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveTemplateDiskSize() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveTemplateDiskSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
