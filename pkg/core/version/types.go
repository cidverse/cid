package version

type ReleaseType int32

const (
	ReleaseNone  ReleaseType = 0
	ReleasePatch ReleaseType = 1
	ReleaseMinor ReleaseType = 2
	ReleaseMajor ReleaseType = 3
)

func HighestReleaseType(numbers []ReleaseType) ReleaseType {
	max := numbers[0]
	for _, value := range numbers {
		if value > max {
			max = value
		}
	}
	return max
}
