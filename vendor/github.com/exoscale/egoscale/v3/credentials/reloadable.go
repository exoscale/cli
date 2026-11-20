package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultReloadInterval is how often NewReloadableCredentials re-reads its Source.
	DefaultReloadInterval = 30 * time.Second
)

// ErrInvalidFormat is returned when the credentials read don't have the format of an Exoscale
// API key and secret.
var ErrInvalidFormat = errors.New("invalid API key or secret format")

var (
	apiKeyFormat    = regexp.MustCompile(`^EXO[0-9a-f]{24}$`)
	apiSecretFormat = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
)

// checkFormat returns ErrInvalidFormat if v isn't an Exoscale API key and secret.
// The values aren't part of the error, which is logged: a key file may hold a secret by mistake.
func checkFormat(v Value) error {
	if !apiKeyFormat.MatchString(v.APIKey) {
		return fmt.Errorf("API key: %w", ErrInvalidFormat)
	}
	if !apiSecretFormat.MatchString(v.APISecret) {
		return fmt.Errorf("API secret: %w", ErrInvalidFormat)
	}

	return nil
}

// A Reloader is a Provider whose credentials can change while the program runs.
// The v3 client calls Reload when the API rejects the key it signed a request with,
// then retries the request once if the credentials changed.
type Reloader interface {
	Provider

	// Reload re-reads the credentials. Changes are reported by the Provider's IsExpired.
	Reload(ctx context.Context) error
}

// A Source reads the credentials from where they are stored, every time it is called.
type Source func() (Value, error)

// KubernetesSecretSource returns a Source reading the API key and secret from the apiKeyFile and
// apiSecretFile files of mountPath, a Kubernetes Secret mounted as a volume: the file names are
// the keys of the Secret.
func KubernetesSecretSource(mountPath, apiKeyFile, apiSecretFile string) Source {
	return func() (Value, error) {
		read := func(name string) (string, error) {
			b, err := os.ReadFile(filepath.Join(mountPath, name))
			if err != nil {
				return "", err
			}

			return strings.TrimSpace(string(b)), nil
		}

		apiKey, err := read(apiKeyFile)
		if err != nil {
			return Value{}, fmt.Errorf("kubernetes secret source: %w", err)
		}
		apiSecret, err := read(apiSecretFile)
		if err != nil {
			return Value{}, fmt.Errorf("kubernetes secret source: %w", err)
		}

		v := Value{APIKey: apiKey, APISecret: apiSecret}
		if !v.IsSet() {
			return Value{}, fmt.Errorf("kubernetes secret source: %s: %w", mountPath, ErrMissingIncomplete)
		}

		return v, nil
	}
}

// JSONFileSource returns a Source reading the API key and secret from a JSON file provided in the
// following format:
//
//	{
//	  "api_key": "...",
//	  "api_secret": "..."
//	}
//
// Any other key found in the JSON object is ignored.
func JSONFileSource(path string) Source {
	return func() (Value, error) {
		b, err := os.ReadFile(path)
		if err != nil {
			return Value{}, fmt.Errorf("json file source: %w", err)
		}

		var file struct {
			APIKey    string `json:"api_key"`
			APISecret string `json:"api_secret"`
		}
		if err := json.Unmarshal(b, &file); err != nil {
			return Value{}, fmt.Errorf("json file source: %s: %w", path, err)
		}

		v := Value{APIKey: strings.TrimSpace(file.APIKey), APISecret: strings.TrimSpace(file.APISecret)}
		if !v.IsSet() {
			return Value{}, fmt.Errorf("json file source: %s: %w", path, ErrMissingIncomplete)
		}

		return v, nil
	}
}

// ReloadableOpt represents a ReloadableProvider option.
type ReloadableOpt func(*ReloadableProvider)

// ReloadableOptWithLogHandler returns a ReloadableOpt logging the reloads, and the errors when
// the Source can't be read or is invalid. Nothing is logged by default.
func ReloadableOptWithLogHandler(h slog.Handler) ReloadableOpt {
	return func(p *ReloadableProvider) {
		p.logger = slog.New(h)
	}
}

// A ReloadableProvider is a Reloader whose credentials are re-read from a Source while the
// program runs, so that a rotated API key is used without restarting it, and without creating
// a new client: the v3 client gets its credentials for every request.
//
// The Source is only re-read by Reload: when the API rejects the key, or by a wrapper adding
// another strategy, like NewPeriodicReloadProvider. New credentials are used as soon as they are
// read. If the Source can't be read, is incomplete or doesn't have the format of an API key and
// secret, the last credentials are kept and the error is logged.
//
// Retrieve doesn't read the Source: it returns the last credentials read.
type ReloadableProvider struct {
	source Source
	logger *slog.Logger

	// reloadMu serializes the reads of the Source.
	reloadMu sync.Mutex
	// lastRead is when the last read of the Source started.
	lastRead time.Time

	mu      sync.RWMutex
	value   Value
	expired bool
}

// NewReloadableProvider returns a ReloadableProvider reading its credentials from source.
// The first read happens here and its error is returned.
func NewReloadableProvider(source Source, opts ...ReloadableOpt) (*ReloadableProvider, error) {
	p := &ReloadableProvider{
		source: source,
		logger: slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(p)
	}

	p.lastRead = time.Now()
	v, err := p.read()
	if err != nil {
		return nil, fmt.Errorf("reloadable provider: %w", err)
	}
	p.value = v
	p.expired = true

	return p, nil
}

// NewReloadableCredentials returns Credentials backed by a ReloadableProvider reading source,
// also re-read every DefaultReloadInterval until ctx is done (see NewPeriodicReloadProvider).
func NewReloadableCredentials(ctx context.Context, source Source, opts ...ReloadableOpt) (*Credentials, error) {
	p, err := NewReloadableProvider(source, opts...)
	if err != nil {
		return nil, err
	}

	return NewCredentials(NewPeriodicReloadProvider(ctx, p, DefaultReloadInterval)), nil
}

// Retrieve returns the last credentials read from the Source.
func (p *ReloadableProvider) Retrieve() (Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.expired = false

	return p.value, nil
}

// IsExpired returns true when new credentials were read since the last Retrieve.
// Credentials.Get calls it on every request: that's how reloaded credentials reach the client.
func (p *ReloadableProvider) IsExpired() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.expired
}

// Reload re-reads the Source, unless a read started after Reload was called: requests rejected
// together wait for the same read instead of each reading the Source again.
// It returns an error if the Source can't be read or is invalid; the previous credentials are
// kept in that case.
func (p *ReloadableProvider) Reload(ctx context.Context) error {
	return p.reload(ctx, time.Now())
}

func (p *ReloadableProvider) reload(ctx context.Context, calledAt time.Time) error {
	p.reloadMu.Lock()
	defer p.reloadMu.Unlock()

	if p.lastRead.After(calledAt) {
		return nil
	}
	p.lastRead = time.Now()

	v, err := p.read()
	if err != nil {
		p.logger.ErrorContext(ctx, "keeping current Exoscale API credentials, unable to read new ones", "error", err)
		return fmt.Errorf("reloadable provider: %w", err)
	}

	p.mu.RLock()
	current := p.value
	p.mu.RUnlock()
	if v == current {
		return nil
	}

	p.mu.Lock()
	p.value = v
	p.expired = true
	p.mu.Unlock()

	p.logger.InfoContext(ctx, "Exoscale API credentials reloaded", "api_key", v.APIKey)

	return nil
}

func (p *ReloadableProvider) read() (Value, error) {
	v, err := p.source()
	if err != nil {
		return Value{}, err
	}

	return v, checkFormat(v)
}

// A PeriodicReloadProvider is a Reloader also reloaded at a regular interval, so that new
// credentials are picked up before the API rejects the old ones.
type PeriodicReloadProvider struct {
	Reloader
}

// NewPeriodicReloadProvider returns a PeriodicReloadProvider calling the Reload of p every
// interval (DefaultReloadInterval if it isn't positive), until ctx is done.
func NewPeriodicReloadProvider(ctx context.Context, p Reloader, interval time.Duration) *PeriodicReloadProvider {
	if interval <= 0 {
		interval = DefaultReloadInterval
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = p.Reload(ctx)
			}
		}
	}()

	return &PeriodicReloadProvider{Reloader: p}
}
