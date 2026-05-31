package constants

import "github.com/cidverse/cid/pkg/util"

var (
	Version          = "0.11.2"
	CommitHash       = "none"
	BuildAt          = "unknown"
	RepositoryStatus = "clean"
	BinaryHash       = ""
)

func init() {
	hash, err := util.GetExecutableHash()
	if err != nil {
		panic(err)
	}
	BinaryHash = hash
}
