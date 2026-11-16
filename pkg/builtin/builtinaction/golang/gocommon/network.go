package gocommon

import "github.com/cidverse/cid/pkg/core/actionsdk"

// Network is read-only access to the go module proxy and checksum database. Module zips are redirected to the
// proxy's bucket on storage.googleapis.com, the rest of that host (any public bucket, incl. uploads) stays blocked.
var Network = []actionsdk.ActionAccessNetwork{
	{Host: "proxy.golang.org:443", HTTPRules: actionsdk.HTTPReadOnly()},
	{Host: "sum.golang.org:443", HTTPRules: actionsdk.HTTPReadOnly("/lookup/**", "/tile/**", "/latest")},
	{Host: "storage.googleapis.com:443", HTTPRules: actionsdk.HTTPReadOnly("/proxy-golang-org-prod/**")},
}
