package transport

import (
	"regexp"
)

type HTTPPathMatcher struct {
	Method  string
	Path    string
	Pattern string
	matcher *regexp.Regexp
}

func NewHTTPPathMatcher(method, path, pattern string) *HTTPPathMatcher {
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		matcher = nil
	}

	return &HTTPPathMatcher{
		Method:  method,
		Path:    path,
		Pattern: pattern,
		matcher: matcher,
	}
}

func (m *HTTPPathMatcher) Match(method string, path string) bool {
	if m.matcher == nil {
		return false
	}
	if m.Method != method {
		return false
	}

	return m.matcher.MatchString(path)
}
