package common

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDetectSBOMFormat(t *testing.T) {
	format, version, ok := DetectSBOMFormat([]byte(`{"bomFormat":"CycloneDX","specVersion":"1.6"}`))
	assert.True(t, ok)
	assert.Equal(t, SBOMFormatCycloneDX, format)
	assert.Equal(t, "1.6", version)

	format, version, ok = DetectSBOMFormat([]byte(`{"spdxVersion":"SPDX-2.3"}`))
	assert.True(t, ok)
	assert.Equal(t, SBOMFormatSPDX, format)
	assert.Equal(t, "2.3", version)

	_, _, ok = DetectSBOMFormat([]byte(`{"name":"not-an-sbom"}`))
	assert.False(t, ok)
	_, _, ok = DetectSBOMFormat([]byte(`not json`))
	assert.False(t, ok)
}

func TestSBOMArtifactName(t *testing.T) {
	assert.Equal(t, "gradle-sub-build-reports-cyclonedx-bom.cdx.json", SBOMArtifactName("gradle", "sub/build/reports/cyclonedx/bom.json", SBOMFormatCycloneDX))
	assert.Equal(t, "maven-target-classes-meta-inf-sbom-application.cdx.json", SBOMArtifactName("maven", "target/classes/META-INF/sbom/application.cdx.json", SBOMFormatCycloneDX))
	assert.Equal(t, "maven-target-app.spdx.json", SBOMArtifactName("maven", "target/app.spdx.json", SBOMFormatSPDX))
}

func TestSBOMProjectDirs(t *testing.T) {
	module := &actionsdk.ProjectModule{
		ModuleDir: "/my-project",
		Submodules: []*actionsdk.ProjectModule{
			{ModuleDir: "/my-project/web/frontend"},
			{ModuleDir: "/my-project/api"},
			{ModuleDir: "/my-project/core", Submodules: []*actionsdk.ProjectModule{{ModuleDir: "/my-project/core/impl"}}},
			{ModuleDir: "/other-project"}, // outside the module
			nil,
		},
	}

	assert.Equal(t, []string{".", "api", "core", "core/impl", "web/frontend"}, SBOMProjectDirs(module))
}

func TestDiscoverSBOMs(t *testing.T) {
	sdk := TestSetup(t)
	module := &actionsdk.ProjectModule{
		ModuleDir:  "/my-project",
		Slug:       "my-module",
		Submodules: []*actionsdk.ProjectModule{{ModuleDir: "/my-project/sub"}},
	}
	discovery := SBOMDiscovery{
		Source: "maven",
		Paths:  []string{"target/bom.json", "target/classes/META-INF/sbom/application.cdx.json", "target/app.spdx.json"},
	}

	sdk.On("FileExistsV1", "/my-project/target/bom.json").Return(true)
	sdk.On("FileExistsV1", "/my-project/target/classes/META-INF/sbom/application.cdx.json").Return(true)
	sdk.On("FileExistsV1", "/my-project/target/app.spdx.json").Return(false)
	sdk.On("FileExistsV1", "/my-project/sub/target/bom.json").Return(true)
	sdk.On("FileExistsV1", "/my-project/sub/target/classes/META-INF/sbom/application.cdx.json").Return(false)
	sdk.On("FileExistsV1", "/my-project/sub/target/app.spdx.json").Return(true)
	sdk.On("FileReadV1", "/my-project/target/bom.json").Return(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`, nil)
	sdk.On("FileReadV1", "/my-project/target/classes/META-INF/sbom/application.cdx.json").Return(`{"bomFormat":"CycloneDX","specVersion":"1.6"}`, nil)
	sdk.On("FileReadV1", "/my-project/sub/target/bom.json").Return(`{"not":"an sbom"}`, nil) // content check
	sdk.On("FileReadV1", "/my-project/sub/target/app.spdx.json").Return(`{"spdxVersion":"SPDX-2.3"}`, nil)
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(r actionsdk.ArtifactUploadRequest) bool {
		return r.File == "maven-target-bom.cdx.json" && r.Module == "my-module" && r.Type == SBOMArtifactType && r.Format == SBOMFormatCycloneDX && r.FormatVersion == "1.5"
	})).Return("", "", nil).Once()
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(r actionsdk.ArtifactUploadRequest) bool {
		return r.File == "maven-target-classes-meta-inf-sbom-application.cdx.json" && r.FormatVersion == "1.6"
	})).Return("", "", nil).Once()
	sdk.On("ArtifactUploadV1", mock.MatchedBy(func(r actionsdk.ArtifactUploadRequest) bool {
		return r.File == "maven-sub-target-app.spdx.json" && r.Format == SBOMFormatSPDX && r.FormatVersion == "2.3"
	})).Return("", "", nil).Once()

	count, err := DiscoverSBOMs(sdk, module, discovery)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}
