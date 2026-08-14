package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ResourcesSuite validates the api.resources.yaml loader.
type ResourcesSuite struct {
	suite.Suite
}

func TestResourcesSuite(t *testing.T) {
	suite.Run(t, new(ResourcesSuite))
}

func (s *ResourcesSuite) fixture(name string) string {
	return filepath.Join("testdata", name)
}

func (s *ResourcesSuite) TestLoadValid() {
	res, err := Load(s.fixture("valid.yaml"))
	s.Require().NoError(err)
	s.Equal(1, res.Version)
	s.Len(res.Resources, 2)

	search, ok := res.Resources["search_laws"]
	s.Require().True(ok, "search_laws resource missing")
	s.Equal("https://nuevo.leychile.cl", search.URL)
	s.Equal("/servicios/buscarjson", search.Path)
	s.Equal("GET", search.Method)
	s.Equal(time.Duration(10*time.Second), time.Duration(search.Timeout))
	s.Equal(3, search.Retry.Attempts)
	s.Equal(time.Duration(500*time.Millisecond), time.Duration(search.Retry.Backoff))
	s.Equal(time.Duration(5*time.Second), time.Duration(search.Retry.MaxBackoff))
	s.Equal(5, search.CircuitBreaker.FailureThreshold)
	s.Equal(2, search.CircuitBreaker.SuccessThreshold)
	s.Equal(time.Duration(30*time.Second), time.Duration(search.CircuitBreaker.ResetTimeout))

	getLaw, ok := res.Resources["get_law"]
	s.Require().True(ok, "get_law resource missing")
	s.Equal("/servicios/Navegar/get_norma_json", getLaw.Path)
}

func (s *ResourcesSuite) TestLoadMalformedYAML() {
	_, err := Load(s.fixture("malformed.yaml"))
	s.Error(err)
	s.Contains(err.Error(), "parse resources file")
}

func (s *ResourcesSuite) TestLoadMissingPath() {
	_, err := Load(s.fixture("missing_path.yaml"))
	s.Error(err)
	s.Contains(err.Error(), `resource "broken"`)
	s.Contains(err.Error(), "path is required")
}

func (s *ResourcesSuite) TestLoadNegativeTimeout() {
	_, err := Load(s.fixture("negative_timeout.yaml"))
	s.Error(err)
	s.Contains(err.Error(), "timeout must be > 0")
}

func (s *ResourcesSuite) TestLoadMissingFile() {
	_, err := Load(s.fixture("does_not_exist.yaml"))
	s.Error(err)
	s.Contains(err.Error(), "read resources file")
}

func (s *ResourcesSuite) TestLoadUnsupportedVersion() {
	_, err := Load(s.fixture("bad_version.yaml"))
	s.Error(err)
	s.Contains(err.Error(), "unsupported version")
}
