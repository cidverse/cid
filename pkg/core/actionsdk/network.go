package actionsdk

// HTTPReadOnly allows downloads only: GET / HEAD on the given paths (none = any path). Use it for package registries and
// other download hosts, it blocks publishing with credentials from the job environment and uploads as an exfiltration
// channel. Implies request inspection (type http), on https the client must trust the egress inspection ca.
func HTTPReadOnly(paths ...string) []ActionAccessNetworkHTTPRule {
	return []ActionAccessNetworkHTTPRule{{Methods: []string{"GET", "HEAD"}, Paths: paths}}
}

// HTTPRule allows the given methods on the given paths (none = any path)
func HTTPRule(methods []string, paths ...string) ActionAccessNetworkHTTPRule {
	return ActionAccessNetworkHTTPRule{Methods: methods, Paths: paths}
}
