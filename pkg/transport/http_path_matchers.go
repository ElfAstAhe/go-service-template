package transport

import (
	"strings"
)

type HTTPPathMatchers struct {
	WatchPaths map[string][]*HTTPPathMatcher
}

func NewHTTPPathMatchers(matchers []*HTTPPathMatcher) *HTTPPathMatchers {
	watchPaths := make(map[string][]*HTTPPathMatcher)

	for _, matcher := range matchers {
		slice, ok := watchPaths[matcher.Method]
		if !ok {
			slice = make([]*HTTPPathMatcher, 0)
		}

		if watchPathExists(slice, matcher.Method, matcher.Path) {
			continue
		}

		watchPaths[matcher.Method] = append(slice, matcher)
	}

	return &HTTPPathMatchers{
		WatchPaths: watchPaths,
	}
}

func (hpm *HTTPPathMatchers) Match(method string, path string) bool {
	return hpm.GetPathMatcher(method, path) != nil
}

func (hpm *HTTPPathMatchers) GetPathMatcher(method string, path string) *HTTPPathMatcher {
	slice, ok := hpm.WatchPaths[method]
	if !ok {
		return nil
	}

	for _, item := range slice {
		if strings.TrimSpace(method) == item.Method && strings.TrimSpace(path) == item.Path {
			return item
		}
	}

	return nil
}

func watchPathExists(src []*HTTPPathMatcher, method string, path string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(method) == "" || len(src) == 0 {
		return false
	}

	for _, item := range src {
		if item.Method == method && item.Path == path {
			return true
		}
	}

	return false
}
