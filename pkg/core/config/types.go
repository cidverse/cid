package config

import (
	"github.com/cidverse/cid/pkg/core/catalog"
)

type ToolBinary struct {
	Binary  string `yaml:"binary"`
	Version string `yaml:"version"`
}

type PathDiscoveryRule struct {
	Binary         []string
	Lookup         []PathDiscoveryRuleLookup
	LookupSuffixes []string `yaml:"lookup-suffixes"`
	Path           string
	ResolvedBinary string
}

// PathDiscoveryRuleLookup is used to discover local tool installations based on ENV vars
type PathDiscoveryRuleLookup struct {
	Key            string   `yaml:"key"`             // env name
	KeyAliases     []string `yaml:"key-aliases"`     // env aliases
	Directory      string   `yaml:"directory"`       // directory
	Version        string   `yaml:"version"`         // version
	VersionCommand string   `yaml:"version-command"` // command to get version
	VersionRegex   string   `yaml:"version-regex"`   // regex to extract version
}

// CIDConfig is the full struct of the configuration file
type CIDConfig struct {
	CommandExecutionTypes []string `yaml:"command-execution-types,omitempty"`

	Paths       PathConfig
	Conventions ProjectConventions
	Env         map[string]string

	// Dependencies holds a key value map of required versions
	Dependencies map[string]string `yaml:"dependencies,omitempty"`

	// LocalTools holds a list to lookup locally installed tools for command execution
	LocalTools []PathDiscoveryRule `yaml:"localtools,omitempty"`

	// CatalogSources
	CatalogSources map[string]*catalog.Source `yaml:"catalog_sources,omitempty"`

	// Registry holding all known images, actions, workflows, ...
	Registry catalog.Config `yaml:"registry,omitempty"`
}
