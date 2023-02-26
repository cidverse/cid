package config

import (
	"github.com/cidverse/cid/pkg/core/catalog"
)

type ToolBinary struct {
	Binary  string `yaml:"binary"`
	Version string `yaml:"version"`
}

type ToolLocal struct {
	Binary         []string
	Lookup         []ToolLocalLookup
	LookupSuffixes []string `yaml:"lookup-suffixes"`
	Path           string
	ResolvedBinary string
}

// ToolLocalLookup is used to discover local tool installations based on ENV vars
type ToolLocalLookup struct {
	Key     string // env name
	Version string // version
}

// CIDConfig is the full stuct of the configuration file
type CIDConfig struct {
	Paths       PathConfig
	Mode        ExecutionType `default:"PREFER_LOCAL"`
	Conventions ProjectConventions
	Env         map[string]string

	// Dependencies holds a key value map of required versions
	Dependencies map[string]string `yaml:"dependencies,omitempty"`

	// LocalTools holds a list to lookup locally installed tools for command execution
	LocalTools []ToolLocal `yaml:"localtools,omitempty"`

	// CatalogSources
	CatalogSources map[string]*catalog.Source `yaml:"catalog_sources,omitempty"`

	// Registry holding all known images, actions, workflows, ...
	Registry catalog.Config `yaml:"registry,omitempty"`
}
