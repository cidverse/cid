package container

import (
	"embed"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

//go:embed dockerfiles/*
var DockerfileFS embed.FS

type Platform struct {
	OS   string
	Arch string
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
	for _, line := range strings.Split(strings.TrimSuffix(dockerfileContent, "\n"), "\n") {
		if strings.HasPrefix(line, "# syntax=") {
			return strings.TrimRight(strings.TrimPrefix(line, "# syntax="), "\r")
		}
	}

	return ""
}

func getDockerfileTargetPlatforms(dockerfileContent string) []Platform {
	var platforms []Platform

	for _, line := range strings.Split(strings.TrimSuffix(dockerfileContent, "\n"), "\n") {
		if strings.HasPrefix(line, "# platforms=") {
			platformsLine := strings.TrimRight(strings.TrimPrefix(line, "# platforms="), "\r")

			for _, element := range strings.Split(platformsLine, ",") {
				elementSections := strings.Split(element, "/")

				if len(elementSections) == 2 {
					platforms = append(platforms, Platform{strings.ToLower(elementSections[0]), strings.ToLower(elementSections[1])})
				} else {
					log.Warn().Str("platform", element).Msg("skipping invalid platform definition from dockerfile")
				}
			}
		}
	}

	if len(platforms) == 0 {
		platforms = append(platforms, Platform{"linux", "amd64"})
	}

	return platforms
}

func getDockerfileTargetImage(dockerfileContent string) string {
	image := ""
	for _, line := range strings.Split(strings.TrimSuffix(dockerfileContent, "\n"), "\n") {
		if strings.HasPrefix(line, "# image=") {
			image = strings.TrimRight(strings.TrimPrefix(line, "# image="), "\r")
		}
	}

	if len(image) > 0 {
		for _, line := range strings.Split(strings.TrimSuffix(dockerfileContent, "\n"), "\n") {
			if strings.Contains(line, "ARG ") && strings.Contains(line, "_VERSION") {
				args := strings.SplitN(line, "=", 2)
				image = image + ":" + strings.TrimRight(args[1], "\r")
				break
			}
		}
	}

	return image
}
