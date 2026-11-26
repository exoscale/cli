package instance

import "testing"

func TestInstanceListLabelsFilter(t *testing.T) {
	tests := []struct {
		name    string
		labels  map[string]string
		want    string
		wantErr bool
	}{
		{name: "no labels", labels: nil, want: ""},
		{name: "one label", labels: map[string]string{"app": "web"}, want: "app=web"},
		{
			name:   "several labels are sorted by key",
			labels: map[string]string{"env": "prod", "app": "web", "team": "nlb"},
			want:   "app=web env=prod team=nlb",
		},
		{name: "empty key", labels: map[string]string{"": "web"}, wantErr: true},
		{name: "empty value", labels: map[string]string{"app": ""}, wantErr: true},
		{name: "space in key", labels: map[string]string{"my app": "web"}, wantErr: true},
		{name: "space in value", labels: map[string]string{"app": "my web"}, wantErr: true},
		{name: "equal sign in value", labels: map[string]string{"app": "a=b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instanceListLabelsFilter(tt.labels)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
