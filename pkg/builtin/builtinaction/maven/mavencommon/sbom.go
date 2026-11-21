package mavencommon

import "github.com/cidverse/cid/pkg/builtin/builtinaction/common"

// SBOMDiscovery lists where Maven SBOM plugins write their output, checked in the module and every submodule (reactor module) directory.
var SBOMDiscovery = common.SBOMDiscovery{
	Source: "maven",
	Paths: []string{
		"target/bom.json", // cyclonedx-maven-plugin (makeBom / makeAggregateBom, default outputName)
		"target/classes/META-INF/sbom/application.cdx.json", // spring boot 3.3+ (spring-boot-starter-parent cyclonedx config)
	},
}
