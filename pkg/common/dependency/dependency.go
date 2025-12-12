package dependency

import (
	"fmt"
)

type Dependency struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	Version    string `json:"version,omitempty"`
	Hash       string `json:"hash,omitempty"`
	Repository string `json:"repository,omitempty"`
}

func (wd Dependency) AsPackageUrl() string {
	packageType := NormalizePackageType(wd.Type)
	base := fmt.Sprintf("pkg:%s/%s", packageType, wd.Id)

	if wd.Type == "docker" {
		version := ""
		if wd.Hash != "" {
			version = "@sha256:" + wd.Hash
		} else if wd.Version != "" {
			version = "@" + wd.Version
		}

		query := ""
		if wd.Repository != "" {
			query = "?repository_url=" + wd.Repository
		}

		return base + version + query
	}

	return base + "@" + wd.Version
}

func (wd Dependency) AsPackageUrlNoVersion() string {
	packageType := NormalizePackageType(wd.Type)
	base := fmt.Sprintf("pkg:%s/%s", packageType, wd.Id)

	if wd.Type == "docker" {
		query := ""
		if wd.Repository != "" {
			query = "?repository_url=" + wd.Repository
		}

		return base + query
	}

	return base
}

// AsDependencyReference returns a string representation of the dependency suitable for use with their respective package managers or tooling.
func (wd Dependency) AsDependencyReference() string {
	if wd.Type == "docker" {
		if wd.Hash != "" {
			return fmt.Sprintf("%s/%s@sha256:%s", wd.Repository, wd.Id, wd.Hash)
		} else {
			return fmt.Sprintf("%s/%s:%s", wd.Repository, wd.Id, wd.Version)
		}
	}

	return fmt.Sprintf("%s:%s", wd.Id, wd.Version)
}

func NormalizePackageType(packageType string) string {
	// see types at https://github.com/package-url/purl-spec/blob/main/purl-types-index.json
	return packageType
}
