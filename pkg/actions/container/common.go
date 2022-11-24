package container

import (
	"embed"
	"github.com/oriser/regroup"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

//go:embed dockerfiles/*
var DockerfileFS embed.FS

type Platform struct {
	OS      string
	Arch    string
	Variant string
}

// Platform returns the platform string (linux/amd64, linux/arm64/v8, ...)
func (p Platform) Platform(separator string) string {
	if p.Variant == "" {
		return p.OS + separator + p.Arch
	}

	return p.OS + separator + p.Arch + separator + p.Variant
}

// DetectAppType checks what kind of app the project is (via artifacts, should run after build actions)
func DetectAppType(ctx *api.ActionExecutionContext) string {
	// java | jar
	files, filesErr := filesystem.FindFilesByExtension(filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact), []string{".jar"})
	if filesErr != nil {
		return ""
	}

	if len(files) > 0 {
		return "jar"
	}

	return ""
}

func getFullImage(host string, repository string, tag string) string {
	if len(host) > 0 {
		return host + "/" + repository + ":" + tag
	}

	return repository + ":" + tag
}

func getDockerfileSyntax(dockerfileContent string) string {
	return parseLine(dockerfileContent, "# syntax=")
}

func getDockerfileTargetPlatforms(dockerfileContent string) []Platform {
	var platforms []Platform

	platformsLine := parseLine(dockerfileContent, "# platforms=")
	for _, element := range strings.Split(platformsLine, ",") {
		elementSections := strings.Split(element, "/")

		if len(elementSections) == 2 {
			platforms = append(platforms, Platform{strings.ToLower(elementSections[0]), strings.ToLower(elementSections[1]), ""})
		} else if len(elementSections) == 3 {
			platforms = append(platforms, Platform{strings.ToLower(elementSections[0]), strings.ToLower(elementSections[1]), strings.ToLower(elementSections[2])})
		} else {
			log.Debug().Str("platform", element).Msg("no platform definition in dockerfile, not building a multi-arch image")
		}
	}

	if len(platforms) == 0 {
		platforms = append(platforms, Platform{"linux", "amd64", ""})
	}

	return platforms
}

func getDockerfileTargetImageWithVersion(dockerfileContent string, suggestedName string) string {
	image := parseLine(dockerfileContent, "# image=")
	tagRegex := parseLine(dockerfileContent, "# tag-regex=")
	ver := ""

	if len(tagRegex) > 0 {
		expr := regroup.MustCompile(tagRegex)

		match, matchErr := expr.Groups(dockerfileContent)
		if matchErr == nil {
			ver = match["major"] + "." + match["minor"] + "." + match["patch"]
			if match["build"] != "" {
				ver = ver + "." + match["build"]
			}
		} else {
			log.Err(matchErr).Msg("failed to match")
		}
	} else if len(image) > 0 {
		for _, line := range strings.Split(strings.TrimSuffix(dockerfileContent, "\n"), "\n") {
			if strings.Contains(line, "ARG ") && strings.Contains(line, "_VERSION") {
				args := strings.SplitN(line, "=", 2)
				ver = strings.TrimRight(args[1], "\r")
				break
			}
		}
	}

	if len(image) > 0 && len(ver) > 0 {
		return image + ":" + ver
	}

	return suggestedName
}

func parseLine(content string, prefix string) string {
	for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimRight(strings.TrimPrefix(line, prefix), "\r")
		}
	}

	return ""
}

func getFirstExistingFile(files []string) string {
	for _, file := range files {
		if filesystem.FileExists(file) {
			return file
		}
	}

	return ""
}
