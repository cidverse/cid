package policy

import (
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"
)

// HTTPConfig contains the request rules of a `type: http` endpoint
type HTTPConfig struct {
	Rules []HTTPRule `yaml:"rules,omitempty"` // allowed requests, no rules = every request allowed
}

func (c HTTPConfig) isZero() bool {
	return len(c.Rules) == 0
}

// HTTPRule allows http requests by method and path, everything that no rule allows is denied.
//
// Methods / Paths empty = any. In paths, `*` matches within one path segment and `**` (as a whole segment) matches one
// or more segments, so `/repos/org/repo/**` covers everything below the repository, but not `/repos/org/repo` itself.
type HTTPRule struct {
	Methods []string `yaml:"methods,omitempty"`
	Paths   []string `yaml:"paths,omitempty"`
}

func (r HTTPRule) Validate() error {
	for _, p := range r.Paths {
		if !strings.HasPrefix(p, "/") {
			return fmt.Errorf("http rule path %q must start with /", p)
		}
		if strings.ContainsAny(p, " \t,|") {
			return fmt.Errorf("http rule path %q must not contain whitespace, ',' or '|'", p) // short form separators
		}
		for _, segment := range strings.Split(p, "/") {
			if _, err := path.Match(segment, ""); err != nil {
				return fmt.Errorf("invalid http rule path %q: %w", p, err)
			}
		}
	}
	for _, m := range r.Methods {
		if m == "" || strings.ContainsAny(m, " \t/,|:") {
			return fmt.Errorf("invalid http rule method %q", m)
		}
	}
	return nil
}

func (r HTTPRule) String() string {
	methods, paths := "*", "/**"
	if len(r.Methods) > 0 {
		methods = strings.Join(r.Methods, ",")
	}
	if len(r.Paths) > 0 {
		paths = strings.Join(r.Paths, ",")
	}
	return methods + " " + paths
}

func (r HTTPRule) matches(method string, requestPath string) bool {
	if len(r.Methods) > 0 && !slices.ContainsFunc(r.Methods, func(m string) bool { return m == "*" || strings.EqualFold(m, method) }) {
		return false
	}
	if len(r.Paths) == 0 {
		return true
	}
	segments := strings.Split(requestPath, "/")
	for _, p := range r.Paths {
		if matchPathSegments(strings.Split(p, "/"), segments) {
			return true
		}
	}
	return false
}

func matchPathSegments(pattern, segments []string) bool {
	if len(pattern) == 0 {
		return len(segments) == 0
	}
	if pattern[0] == "**" {
		for i := 1; i <= len(segments); i++ {
			if matchPathSegments(pattern[1:], segments[i:]) {
				return true
			}
		}
		return false
	}
	if len(segments) == 0 {
		return false
	}
	if ok, _ := path.Match(pattern[0], segments[0]); !ok {
		return false
	}
	return matchPathSegments(pattern[1:], segments[1:])
}

// allowsRequest reports whether an http endpoint's rules permit the request
func (e Endpoint) allowsRequest(method string, requestPath string) bool {
	if len(e.HTTP.Rules) == 0 {
		return true
	}
	for _, r := range e.HTTP.Rules {
		if r.matches(method, requestPath) {
			return true
		}
	}
	return false
}

type HTTPDecision struct {
	Allowed bool
	Rule    string // the type http endpoint that allowed the request (short form), empty without inspection
	Reason  string // why the request was denied
}

// EvaluateHTTP checks a request against all endpoints that matched the connection: it is allowed if any `type: http`
// endpoint's rules allow it. Endpoints of other types for the same host add no request access. Rules are matched on the
// decoded path, requests with `.` / `..` segments are always denied. Whether a denied request is blocked or only
// reported depends on the filter mode.
func EvaluateHTTP(endpoints []Endpoint, method string, u *url.URL) HTTPDecision {
	var inspected []Endpoint
	for _, e := range endpoints {
		if e.Type == TypeHTTP {
			inspected = append(inspected, e)
		}
	}
	if len(inspected) == 0 {
		return HTTPDecision{Allowed: true}
	}

	requestPath := u.Path
	if requestPath == "" {
		requestPath = "/"
	}
	if hasDotSegment(requestPath) {
		return HTTPDecision{Reason: "dot segment in path"}
	}
	for _, e := range inspected {
		if e.allowsRequest(method, requestPath) {
			return HTTPDecision{Allowed: true, Rule: e.Rule()}
		}
	}
	return HTTPDecision{Reason: "no rule allows the request"}
}

// Inspects reports whether any endpoint inspects http requests
func Inspects(endpoints []Endpoint) bool {
	for _, e := range endpoints {
		if e.Type == TypeHTTP {
			return true
		}
	}
	return false
}

func hasDotSegment(p string) bool {
	for _, segment := range strings.Split(p, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}
