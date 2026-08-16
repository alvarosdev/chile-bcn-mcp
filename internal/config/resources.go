// Package config loads the API resources contract (api.resources.yaml):
// where and how to call each external endpoint, without hardcoded URLs
// in the code. Loaded once at startup, validated fail-fast.
// The contract is baked into the binary via go:embed (internal/config/api.resources.yaml);
// no external file is required at runtime. Load(path) is kept only for tests with fixtures.
package config

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed api.resources.yaml
var rawResources []byte

// Resources is the root of the api.resources.yaml contract.
type Resources struct {
	Version   int                 `yaml:"version"`
	Resources map[string]Resource `yaml:"resources"`
}

// Resource declares where and how to call one external endpoint.
// The map key is the resource id used by the code (e.g. "search_laws").
type Resource struct {
	URL            string         `yaml:"url"`
	Path           string         `yaml:"path"`
	Method         string         `yaml:"method"`
	Timeout        Duration       `yaml:"timeout"`
	Retry          Retry          `yaml:"retry"`
	CircuitBreaker CircuitBreaker `yaml:"circuit_breaker"`
}

// Retry configures the retry behavior for a resource (count-based).
type Retry struct {
	Attempts   int      `yaml:"attempts"`
	Backoff    Duration `yaml:"backoff"`
	MaxBackoff Duration `yaml:"max_backoff"`
}

// CircuitBreaker configures the count-based circuit breaker of resty:
// it trips when FailureThreshold failures accumulate in the sliding
// window, opens for ResetTimeout, and closes after SuccessThreshold
// consecutive probe successes.
type CircuitBreaker struct {
	FailureThreshold int      `yaml:"failure_threshold"`
	SuccessThreshold int      `yaml:"success_threshold"`
	ResetTimeout     Duration `yaml:"reset_timeout"`
}

// Duration is a time.Duration that unmarshals from strings like "10s".
type Duration time.Duration

// UnmarshalYAML parses a duration string ("10s", "500ms") into Duration.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(dur)
	return nil
}

// LoadEmbedded reads and validates the embedded api.resources.yaml contract.
// The file is baked into the binary via go:embed; no external file is required.
func LoadEmbedded() (*Resources, error) {
	var res Resources
	if err := yaml.Unmarshal(rawResources, &res); err != nil {
		return nil, fmt.Errorf("parse embedded resources: %w", err)
	}
	if err := res.validate(); err != nil {
		return nil, err
	}
	return &res, nil
}

// Load reads and validates the resources contract from path.
// Kept for tests with testdata fixtures; production code should use LoadEmbedded.
func Load(path string) (*Resources, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read resources file: %w", err)
	}
	var res Resources
	if err := yaml.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("parse resources file: %w", err)
	}
	if err := res.validate(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Resources) validate() error {
	if r.Version != 1 {
		return fmt.Errorf("unsupported version %d (want 1)", r.Version)
	}
	if len(r.Resources) == 0 {
		return fmt.Errorf("no resources declared")
	}
	for name, res := range r.Resources {
		if err := res.validate(); err != nil {
			return fmt.Errorf("resource %q: %w", name, err)
		}
	}
	return nil
}

func (r *Resource) validate() error {
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if r.Method != "GET" && r.Method != "POST" {
		return fmt.Errorf("method %q is not supported (GET or POST)", r.Method)
	}
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be > 0")
	}
	if r.Retry.Attempts < 1 {
		return fmt.Errorf("retry.attempts must be >= 1")
	}
	if r.Retry.Backoff <= 0 {
		return fmt.Errorf("retry.backoff must be > 0")
	}
	if r.Retry.MaxBackoff < r.Retry.Backoff {
		return fmt.Errorf("retry.max_backoff must be >= retry.backoff")
	}
	if r.CircuitBreaker.FailureThreshold < 1 || r.CircuitBreaker.SuccessThreshold < 1 {
		return fmt.Errorf("circuit_breaker thresholds must be >= 1")
	}
	if r.CircuitBreaker.ResetTimeout <= 0 {
		return fmt.Errorf("circuit_breaker.reset_timeout must be > 0")
	}
	return nil
}
