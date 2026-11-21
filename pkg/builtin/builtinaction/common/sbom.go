package common

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const (
	SBOMArtifactType    = "sbom"
	SBOMFormatCycloneDX = "cyclonedx-json"
	SBOMFormatSPDX      = "spdx-json"
)

// SBOMOutputArtifacts are the artifact types declared by actions that produce or discover SBOMs.
var SBOMOutputArtifacts = []actionsdk.ActionArtifactType{
	{Type: SBOMArtifactType, Format: SBOMFormatCycloneDX},
	{Type: SBOMArtifactType, Format: SBOMFormatSPDX},
}

// SBOMDiscovery describes where a build system's plugins write SBOMs.
type SBOMDiscovery struct {
	// Source names the discovering action in artifact names, e.g. "maven" -> "maven-target-bom.cdx.json".
	Source string
	// Paths are the SBOM locations relative to each project directory, e.g. "target/bom.json".
	Paths []string
}

var sbomNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// DetectSBOMFormat inspects JSON content and returns the SBOM format and spec version (CycloneDX or SPDX JSON).
func DetectSBOMFormat(content []byte) (format string, version string, ok bool) {
	var doc struct {
		BOMFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
		SPDXVersion string `json:"spdxVersion"`
	}
	if err := json.Unmarshal(content, &doc); err != nil {
		return "", "", false
	}

	if doc.BOMFormat == "CycloneDX" {
		return SBOMFormatCycloneDX, doc.SpecVersion, true
	}
	if strings.HasPrefix(doc.SPDXVersion, "SPDX-") {
		return SBOMFormatSPDX, strings.TrimPrefix(doc.SPDXVersion, "SPDX-"), true
	}

	return "", "", false
}

// SBOMArtifactName returns a unique artifact name for an SBOM, as all SBOMs of a step share one artifact directory.
// e.g. source "gradle", rel "sub/build/reports/cyclonedx/bom.json" -> "gradle-sub-build-reports-cyclonedx-bom.cdx.json"
func SBOMArtifactName(source string, rel string, format string) string {
	base := strings.ToLower(rel)
	for _, suffix := range []string{".cdx.json", ".spdx.json", ".json"} {
		if strings.HasSuffix(base, suffix) {
			base = strings.TrimSuffix(base, suffix)
			break
		}
	}
	base = strings.Trim(sbomNameSanitizer.ReplaceAllString(strings.ReplaceAll(base, "/", "-"), "-"), "-")

	ext := ".cdx.json"
	if format == SBOMFormatSPDX {
		ext = ".spdx.json"
	}

	return source + "-" + base + ext
}

// SBOMProjectDirs returns the module directory and the directories of all its submodules, relative to the module
// directory ("." for the module itself). Repository analysis registers every nested build file (pom.xml,
// build.gradle[.kts], ...) as a submodule of the enclosing module, so this covers all subprojects / reactor modules
// without walking the file tree.
func SBOMProjectDirs(module *actionsdk.ProjectModule) []string {
	dirs := []string{"."}

	var walk func(modules []*actionsdk.ProjectModule)
	walk = func(modules []*actionsdk.ProjectModule) {
		for _, sub := range modules {
			if sub == nil {
				continue
			}
			if rel, err := filepath.Rel(module.ModuleDir, sub.ModuleDir); err == nil && sub.ModuleDir != "" {
				rel = filepath.ToSlash(rel)
				if rel != ".." && !strings.HasPrefix(rel, "../") && !slices.Contains(dirs, rel) {
					dirs = append(dirs, rel)
				}
			}
			walk(sub.Submodules)
		}
	}
	walk(module.Submodules)

	slices.Sort(dirs[1:])
	return dirs
}

// DiscoverSBOMs checks the known SBOM locations of the build system's plugins in the module and every submodule directory
// and stores each SBOM found as an artifact of type `sbom`. Files are verified by content (CycloneDX / SPDX JSON).
// Several SBOMs per module are expected (e.g. per subproject and aggregated) and all of them are kept.
func DiscoverSBOMs(sdk actionsdk.SDKClient, module *actionsdk.ProjectModule, discovery SBOMDiscovery) (int, error) {
	count := 0
	for _, dir := range SBOMProjectDirs(module) {
		for _, candidate := range discovery.Paths {
			rel := path.Join(dir, candidate)
			file := actionsdk.JoinPath(module.ModuleDir, rel)
			if !sdk.FileExistsV1(file) {
				continue
			}

			content, err := sdk.FileReadV1(file)
			if err != nil {
				_ = sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "failed to read sbom", Context: map[string]interface{}{"file": file, "error": err.Error()}})
				continue
			}

			format, version, ok := DetectSBOMFormat([]byte(content))
			if !ok {
				_ = sdk.LogV1(actionsdk.LogV1Request{Level: "warn", Message: "file at sbom location is not a CycloneDX / SPDX json document, skipping", Context: map[string]interface{}{"file": file}})
				continue
			}

			_, _, err = sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:          SBOMArtifactName(discovery.Source, rel, format),
				ContentBytes:  []byte(content),
				Module:        module.Slug,
				Type:          SBOMArtifactType,
				Format:        format,
				FormatVersion: version,
			})
			if err != nil {
				return count, fmt.Errorf("failed to store sbom %s: %w", file, err)
			}
			_ = sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "discovered sbom", Context: map[string]interface{}{"file": file, "format": format, "format_version": version}})
			count++
		}
	}

	return count, nil
}
