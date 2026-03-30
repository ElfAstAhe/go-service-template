package transport

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPMs_GetPM_CorrectData_AllCases(t *testing.T) {
	// prepare
	pms := buildPMS()
	// act
	t.Run("exists GET /", func(t *testing.T) {
		// assert
		assert.NotNil(t, pms.GetPathMatcher(http.MethodGet, "/"))
	})
	t.Run("not exists GET /123", func(t *testing.T) {
		// assert
		assert.Nil(t, pms.GetPathMatcher(http.MethodGet, "/123"))
	})
	t.Run("exists POST /", func(t *testing.T) {
		// assert
		assert.NotNil(t, pms.GetPathMatcher(http.MethodPost, "/"))
	})
	t.Run("not exists POST /123", func(t *testing.T) {
		// assert
		assert.Nil(t, pms.GetPathMatcher(http.MethodPost, "/123"))
	})
	t.Run("not exists empty method", func(t *testing.T) {
		assert.Nil(t, pms.GetPathMatcher(http.MethodDelete, "/"))
	})
}

func TestPMs_Match_CorrectData_AllCases(t *testing.T) {
	// prepare
	pms := buildPMS()
	// act
	t.Run("exists GET /", func(t *testing.T) {
		// assert
		assert.True(t, pms.Match(http.MethodGet, "/"))
	})
	t.Run("not exists GET /123", func(t *testing.T) {
		// assert
		assert.False(t, pms.Match(http.MethodGet, "/123"))
	})
	t.Run("exists POST /", func(t *testing.T) {
		assert.True(t, pms.Match(http.MethodPost, "/"))
	})
	t.Run("not exists POST /123", func(t *testing.T) {
		assert.False(t, pms.Match(http.MethodPost, "/123"))
	})
}

func buildPMS() *HTTPPathMatchers {
	pm1 := NewHTTPPathMatcher(http.MethodGet, "/", `^/[^/]*$`)
	pm2 := NewHTTPPathMatcher(http.MethodPost, "/", `/$`)
	return NewHTTPPathMatchers([]*HTTPPathMatcher{pm1, pm2})
}
