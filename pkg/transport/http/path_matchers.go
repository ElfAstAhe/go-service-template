package http

import (
	"strings"
)

type PathMatchers struct {
	WatchPaths map[string][]*PathMatcher
}

func NewHTTPPathMatchers(matchers []*PathMatcher) *PathMatchers {
	watchPaths := make(map[string][]*PathMatcher)

	for _, matcher := range matchers {
		slice, ok := watchPaths[matcher.Method]
		if !ok {
			slice = make([]*PathMatcher, 0)
		}

		if watchPathExists(slice, matcher.Method, matcher.Path) {
			continue
		}

		watchPaths[matcher.Method] = append(slice, matcher)
	}

	return &PathMatchers{
		WatchPaths: watchPaths,
	}
}

func (hpm *PathMatchers) Match(method string, path string) bool {
	pms, ok := hpm.WatchPaths[method]
	if !ok {
		return false
	}
	for _, pm := range pms {
		if pm.Match(method, path) {
			return true
		}
	}

	return false
}

func (hpm *PathMatchers) GetPathMatcher(method string, path string) *PathMatcher {
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

func watchPathExists(src []*PathMatcher, method string, path string) bool {
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
