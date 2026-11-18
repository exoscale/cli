package storage

import "testing"

func Test_parsePath(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantBucket string
		wantKey    string
		wantErr    string
	}{
		{name: "bucket only", arg: "sos://b", wantBucket: "b"},
		{name: "bucket with trailing slash", arg: "sos://b/", wantBucket: "b"},
		{name: "object", arg: "sos://b/k", wantBucket: "b", wantKey: "k"},
		{name: "nested object", arg: "sos://b/dir/k", wantBucket: "b", wantKey: "dir/k"},
		{name: "prefix keeps its trailing slash", arg: "sos://b/dir/", wantBucket: "b", wantKey: "dir/"},
		{name: "no sos:// prefix", arg: "b/k", wantBucket: "b", wantKey: "k"},
		{name: "empty", arg: "", wantErr: `invalid argument: ""`},
		{name: "prefix only", arg: "sos://", wantErr: `invalid argument: "sos://"`},
		{name: "empty bucket with key", arg: "sos:///k", wantErr: `invalid argument: "sos:///k"`},
		{name: "leading slash", arg: "/k", wantErr: `invalid argument: "/k"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := parsePath(tt.arg)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("ParsePath(%q) error = %v, want %q", tt.arg, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParsePath(%q) unexpected error: %v", tt.arg, err)
			}
			if bucket != tt.wantBucket || key != tt.wantKey {
				t.Errorf("ParsePath(%q) = (%q, %q), want (%q, %q)",
					tt.arg, bucket, key, tt.wantBucket, tt.wantKey)
			}
		})
	}
}
