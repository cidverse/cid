package config

type ToolBinary struct {
	Binary  string `json:"binary"`
	Version string `json:"version"`
}

type ToolContainerImage struct {
	Provides []ToolBinary            `json:"provides"`
	Image    string                  `json:"image"`
	Cache    map[string]ToolCacheDir `json:"cache"`
	User     string                  `json:"user"`
}

type ToolLocal struct {
	Binary         []string
	Lookup         []ToolLocalLookup
	LookupSuffixes []string `json:"lookup-suffixes"`
	Path           string
	ResolvedBinary string
}

// ToolLocalLookup is used to discover local tool installations based on ENV vars
type ToolLocalLookup struct {
	Key     string // env name
	Version string // version
}

type ToolCacheDir struct {
	Id            string
	ContainerPath string `yaml:"dir"`
	MountType     string `yaml:"type"`
}

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact       string `default:"dist"`
	ModuleArtifact string `default:"dist"`
	Temp           string `default:"tmp"`
	Cache          string `default:""`
}

// CIDConfig is the full stuct of the configuration file
type CIDConfig struct {
	Catalog      Catalog `yaml:"catalog,omitempty"`
	Paths        PathConfig
	Mode         ExecutionType `default:"PREFER_LOCAL"`
	Conventions  ProjectConventions
	Env          map[string]string
	Stages       []WorkflowStage             `yaml:"stages"`
	Actions      map[string][]WorkflowAction `yaml:"actions"`
	Dependencies map[string]string

	// LocalTools holds a list to lookup locally installed tools for command execution
	LocalTools []ToolLocal `yaml:"localtools,omitempty"`

	// ContainerImages holds a list of images that provide tools for command execution
	ContainerImages []ToolContainerImage `yaml:"containerimages"`
}
