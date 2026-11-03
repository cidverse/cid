package mavencommon

import (
	"fmt"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

// WrapperJar returns the path of the maven wrapper jar in the given module directory
func WrapperJar(moduleDir string) string {
	return actionsdk.JoinPath(moduleDir, ".mvn", "wrapper", "maven-wrapper.jar")
}

// MavenWrapperCommand returns the maven command, invoking the wrapper jar via java if present (allows execution in plain jdk images)
func MavenWrapperCommand(isUsingWrapper bool, moduleDir string, args string) string {
	if isUsingWrapper {
		return fmt.Sprintf("java -Dmaven.multiModuleProjectDirectory=%q -classpath %q org.apache.maven.wrapper.MavenWrapperMain %s", moduleDir, WrapperJar(moduleDir), args)
	}

	return fmt.Sprintf("mvn %s", args)
}
