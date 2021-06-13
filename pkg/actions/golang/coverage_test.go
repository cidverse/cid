package golang

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCoverageProfile(t *testing.T) {
	cmdOutput := `github.com/cidverse/cid/app.go:24:                                              init                            84.6%
github.com/cidverse/cid/app.go:54:                                              main                            0.0%
github.com/cidverse/cid/pkg/actions/changelog/engine.go:15:                     PreprocessCommits               95.0%
github.com/cidverse/cid/pkg/actions/changelog/engine.go:62:                     ProcessCommits                  100.0%
github.com/cidverse/cid/pkg/actions/changelog/engine.go:116:                    RenderTemplate                  75.0%
github.com/cidverse/cid/pkg/actions/changelog/repo-changelog-generate.go:17:    GetDetails                      100.0%
github.com/cidverse/cid/pkg/actions/changelog/repo-changelog-generate.go:27:    Check                           0.0%
github.com/cidverse/cid/pkg/actions/changelog/repo-changelog-generate.go:32:    Execute                         0.0%
github.com/cidverse/cid/pkg/actions/changelog/repo-changelog-generate.go:86:    init                            100.0%
github.com/cidverse/cid/pkg/actions/changelog/util.go:12:                       GetFileContent                  0.0%
github.com/cidverse/cid/pkg/actions/changelog/util.go:31:                       AddLinks                        100.0%
github.com/cidverse/cid/pkg/actions/upx/upx-optimize.go:14:                     GetDetails                      100.0%
github.com/cidverse/cid/pkg/actions/upx/upx-optimize.go:24:                     Check                           100.0%
github.com/cidverse/cid/pkg/actions/upx/upx-optimize.go:30:                     Execute                         0.0%
github.com/cidverse/cid/pkg/actions/upx/upx-optimize.go:46:                     init                            100.0%
github.com/cidverse/cid/pkg/common/commitanalyser/common.go:11:                 DeterminateNextReleaseVersion   85.0%
github.com/cidverse/cid/pkg/common/commitanalyser/common.go:73:                 getHighestReleaseType           100.0%
total:                                                                          (statements)                    62.7%`

	report := ParseCoverageProfile(cmdOutput)

	assert.Equal(t, "go", report.Language)
	assert.Equal(t, 62.7, report.Percent)
}
