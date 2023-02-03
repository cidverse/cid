package version

type ReleaseType int32

const (
	ReleaseNone  ReleaseType = 0
	ReleasePatch ReleaseType = 1
	ReleaseMinor ReleaseType = 2
	ReleaseMajor ReleaseType = 3
)
