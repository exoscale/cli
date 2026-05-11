package cmd

import (
	"testing"

	"github.com/exoscale/cli/pkg/account"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestResolve(t *testing.T) {
	type want struct {
		key           string
		secret        string
		zone          string
		endpoint      string
		environment   string
		sosEndpoint   string
		clientTimeout int
		secretCmd     []string
	}

	tests := []struct {
		name string
		env  envSources
		file *account.Account
		want want
	}{
		{
			name: "defaults when no sources set",
			want: want{
				zone:        DefaultZone,
				environment: DefaultEnvironment,
				sosEndpoint: DefaultSosEndpoint,
			},
		},
		{
			name: "file profile preserved",
			file: &account.Account{Name: "prod", Key: "file-key", Secret: "file-secret", DefaultZone: "de-fra-1"},
			want: want{key: "file-key", secret: "file-secret", zone: "de-fra-1"},
		},
		{
			name: "env credentials override file",
			env:  envSources{apiKey: "env-key", apiSecret: "env-secret"},
			file: &account.Account{Key: "file-key", Secret: "file-secret"},
			want: want{key: "env-key", secret: "env-secret"},
		},
		{
			name: "env credentials clear secret command",
			env:  envSources{apiKey: "k", apiSecret: "s"},
			file: &account.Account{SecretCommand: []string{"gpg", "--decrypt", "secret.gpg"}},
			want: want{key: "k", secret: "s", secretCmd: nil},
		},
		{
			name: "partial env credentials ignored",
			env:  envSources{apiKey: "env-key-only"},
			file: &account.Account{Key: "file-key", Secret: "file-secret"},
			want: want{key: "file-key", secret: "file-secret"},
		},
		{
			name: "env zone overrides file",
			env:  envSources{zone: "ch-gva-2"},
			file: &account.Account{DefaultZone: "de-fra-1"},
			want: want{zone: "ch-gva-2"},
		},
		{
			name: "file zone preserved when no env zone",
			file: &account.Account{DefaultZone: "de-fra-1"},
			want: want{zone: "de-fra-1"},
		},
		{
			name: "sos endpoint trailing slash stripped",
			env:  envSources{apiKey: "k", apiSecret: "s", sosEndpoint: "https://sos.example.com/"},
			want: want{key: "k", secret: "s", sosEndpoint: "https://sos.example.com"},
		},
		{
			name: "client timeout from env",
			env:  envSources{clientTimeout: ptr(42)},
			want: want{clientTimeout: 42},
		},
		{
			name: "env endpoint overrides file",
			env:  envSources{apiEndpoint: "https://env-endpoint.exo.io"},
			file: &account.Account{Endpoint: "https://file-endpoint.exo.io"},
			want: want{endpoint: "https://env-endpoint.exo.io"},
		},
		{
			name: "env only no file",
			env:  envSources{apiKey: "env-key", apiSecret: "env-secret", zone: "at-vie-1"},
			want: want{key: "env-key", secret: "env-secret", zone: "at-vie-1", environment: DefaultEnvironment},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := fileSources{profile: tc.file}
			acc := resolve(tc.env, fs)

			if tc.want.key != "" {
				require.Equal(t, tc.want.key, acc.Key)
			}
			if tc.want.secret != "" {
				require.Equal(t, tc.want.secret, acc.Secret)
			}
			if tc.want.zone != "" {
				require.Equal(t, tc.want.zone, acc.DefaultZone)
			}
			if tc.want.endpoint != "" {
				require.Equal(t, tc.want.endpoint, acc.Endpoint)
			}
			if tc.want.environment != "" {
				require.Equal(t, tc.want.environment, acc.Environment)
			}
			if tc.want.sosEndpoint != "" {
				require.Equal(t, tc.want.sosEndpoint, acc.SosEndpoint)
			}
			if tc.want.clientTimeout != 0 {
				require.Equal(t, tc.want.clientTimeout, acc.ClientTimeout)
			}
			require.Equal(t, tc.want.secretCmd, acc.SecretCommand)
		})
	}
}
