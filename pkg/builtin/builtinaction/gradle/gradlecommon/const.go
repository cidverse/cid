package gradlecommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
)

var (
	NetworkJvm = []cidsdk.ActionAccessNetwork{
		{Host: "repo.maven.apache.org:443"},
		{Host: "repo1.maven.org:443"},
		{Host: "kotlinlang.org:443"},
	}

	NetworkGradle = []cidsdk.ActionAccessNetwork{
		{Host: "plugins.gradle.org:443"},
		{Host: "plugins-artifacts.gradle.org:443"},
		{Host: "services.gradle.org:443"},
		{Host: "downloads.gradle.org:443"},
		{Host: "jcenter.bintray.com:443"},
	}

	NetworkPublish = []cidsdk.ActionAccessNetwork{
		{Host: "oss.sonatype.org:443"},     // lagcy ossrh
		{Host: "s01.oss.sonatype.org:443"}, // lagcy ossrh
		{Host: "central.sonatype.com:443"}, // new mavenCentral
		{Host: "maven.pkg.github.com:443"},
	}
)
