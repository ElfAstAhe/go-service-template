package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMatcher_CorrectData_AllCases(t *testing.T) {
	// prepare
	pm := NewPathMatcher(http.MethodGet, "/", `^/[^/]*$`)
	// act
	t.Run("match", func(t *testing.T) {
		// assert
		assert.True(t, pm.Match(http.MethodGet, "/123"))
	})
	t.Run("no match", func(t *testing.T) {
		assert.False(t, pm.Match(http.MethodGet, "/123/123"))
	})
}

func TestPathMatcher_IncorrectData_AllCases(t *testing.T) {
	// prepare
	pm := NewPathMatcher(http.MethodGet, "/", `^/[^/]*$`)
	// act
	t.Run("not match empty method", func(t *testing.T) {
		assert.False(t, pm.Match("", "/123"))
	})
	t.Run("not match empty path", func(t *testing.T) {
		assert.False(t, pm.Match(http.MethodGet, ""))
	})
}
