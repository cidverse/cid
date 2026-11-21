package gradlecommon

import "github.com/cidverse/cid/pkg/builtin/builtinaction/common"

// SBOMDiscovery lists where Gradle SBOM plugins write their output, checked in the module and every submodule (subproject) directory.
var SBOMDiscovery = common.SBOMDiscovery{
	Source: "gradle",
	Paths: []string{
		"build/reports/bom.json",                       // org.cyclonedx.bom < 3 (cyclonedxBom)
		"build/reports/cyclonedx/bom.json",             // org.cyclonedx.bom >= 3, aggregate (cyclonedxBom)
		"build/reports/cyclonedx-direct/bom.json",      // org.cyclonedx.bom >= 3, per project (cyclonedxDirectBom)
		"build/reports/application.cdx.json",           // spring boot 3.3+ with org.cyclonedx.bom < 3
		"build/reports/cyclonedx/application.cdx.json", // spring boot with org.cyclonedx.bom >= 3
	},
}
