package mavencommon

import (
	"fmt"
)

const wrapperJar = ".mvn/wrapper/maven-wrapper.jar"

func MavenWrapperCommand(isUsingWrapper bool, args string) string {
	if isUsingWrapper {
		return fmt.Sprintf("java -classpath=%q org.apache.maven.wrapper.MavenWrapperMain %s", wrapperJar, args)
	}

	return fmt.Sprintf("mvn %s", args)
}
