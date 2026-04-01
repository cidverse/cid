package gradlecommon

import "github.com/cidverse/cid/pkg/core/actionsdk"

var (
	NetworkJvm = []actionsdk.ActionAccessNetwork{
		{Host: "repo.maven.apache.org:443"},
		{Host: "repo1.maven.org:443"},
		{Host: "kotlinlang.org:443"},
	}

	NetworkGradle = []actionsdk.ActionAccessNetwork{
		{Host: "plugins.gradle.org:443"},
		{Host: "plugins-artifacts.gradle.org:443"},
		{Host: "services.gradle.org:443"}, // gradle wrapper checksum verification
		{Host: "downloads.gradle.org:443"},
		{Host: "jcenter.bintray.com:443"},
	}

	NetworkPublish = []actionsdk.ActionAccessNetwork{
		{Host: "central.sonatype.com:443"}, // mavenCentral
		{Host: "maven.pkg.github.com:443"},
	}
)
