package suite

import (
	"context"
	"testing"

	"github.com/wasilwamark/vps-init/tests/integration/helpers"
)

type TestSuite struct {
	T             *testing.T
	Ctx           context.Context
	Cleanup       func()
	TestContainer interface{}
	SSHConfig     *helpers.SSHConfig
}

func NewTestSuite(t *testing.T) *TestSuite {
	return &TestSuite{
		T:   t,
		Ctx: context.Background(),
	}
}

func (s *TestSuite) Setup() error {
	return nil
}

func (s *TestSuite) Teardown() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func (s *TestSuite) AssertEqual(expected, actual interface{}, msg string) {
	s.Helper()
	if expected != actual {
		s.T.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func (s *TestSuite) AssertTrue(condition bool, msg string) {
	s.Helper()
	if !condition {
		s.T.Errorf("%s: expected true, got false", msg)
	}
}

func (s *TestSuite) AssertNoError(err error, msg string) {
	s.Helper()
	if err != nil {
		s.T.Errorf("%s: %v", msg, err)
	}
}

func (s *TestSuite) AssertContains(output, substring string) {
	s.Helper()
	if !contains(output, substring) {
		s.T.Errorf("expected output to contain %q, got: %s", substring, output)
	}
}

func (s *TestSuite) AssertContainsError(output, substring string) {
	s.Helper()
	if !contains(output, substring) {
		s.T.Errorf("expected error to contain %q, got: %s", substring, output)
	}
}

func (s *TestSuite) Helper() {
	s.T.Helper()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
