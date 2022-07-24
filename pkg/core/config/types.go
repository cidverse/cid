package config

type ToolBinary struct {
	Binary  string `yaml:"binary"`
	Version string `yaml:"version"`
}

type ToolSecurity struct {
	Capabilities []string `yaml:"capabilities"`
}

type ToolContainerImage struct {
	Provides []ToolBinary   `yaml:"provides"`
	Image    string         `yaml:"image"`
	Cache    []ToolCacheDir `yaml:"cache"`
	Security ToolSecurity   `yaml:"security"`
	User     string         `yaml:"user"`
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

type ToolCacheDir struct {
	ID            string
	ContainerPath string `yaml:"dir"`
	MountType     string `yaml:"type"`
}

// CIDConfig is the full stuct of the configuration file
type CIDConfig struct {
	Paths       PathConfig
	Mode        ExecutionType `default:"PREFER_LOCAL"`
	Conventions ProjectConventions
	Env         map[string]string

	// Catalog holds all currently known actions
	Catalog Catalog `yaml:"catalog,omitempty"`

	// Workflows holds all available workflows
	Workflows []Workflow `yaml:"workflows,omitempty"`

	// Dependencies holds a key value map of required versions
	Dependencies map[string]string `yaml:"dependencies,omitempty"`

	// LocalTools holds a list to lookup locally installed tools for command execution
	LocalTools []ToolLocal `yaml:"localtools,omitempty"`

	// ContainerImages holds a list of images that provide tools for command execution
	ContainerImages []ToolContainerImage `yaml:"containerimages,omitempty"`
}
