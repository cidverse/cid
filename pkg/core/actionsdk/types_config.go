package actionsdk

// ProjectDependency defines model for ProjectDependency.
type ProjectDependency struct {
	Id      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

// ProjectModule defines model for ProjectModule.
type ProjectModule struct {
	ProjectDir            string                   `json:"project_dir,omitempty"`            // ProjectDir project root directory
	ModuleDir             string                   `json:"module_dir,omitempty"`             // ModuleDir module root directory
	Discovery             []ProjectModuleDiscovery `json:"discovery,omitempty"`              // Discovery module detected based on
	Name                  string                   `json:"name,omitempty"`                   // Name module name
	Slug                  string                   `json:"slug,omitempty"`                   // Slug module name
	Type                  string                   `json:"type,omitempty"`                   // Type is the module type
	BuildSystem           string                   `json:"build_system,omitempty"`           // BuildSystem module name
	BuildSystemSyntax     string                   `json:"build_system_syntax,omitempty"`    // BuildSystemSyntax module name
	SpecificationType     string                   `json:"specification_type,omitempty"`     // SpecificationType is the type of the specification
	ConfigType            string                   `json:"config_type,omitempty"`            // ConfigType is the type of the configuration
	DeploymentSpec        string                   `json:"deployment_spec,omitempty"`        // DeploymentSpec is the kind of deployment specification
	DeploymentType        string                   `json:"deployment_type,omitempty"`        // DeploymentType is the type of the deployment
	DeploymentEnvironment string                   `json:"deployment_environment,omitempty"` // DeploymentEnvironment is the environment the deployment is for, e.g. staging, production, ...
	Language              map[string]string        `json:"language,omitempty"`               // Language module name
	Dependencies          []*ProjectDependency     `json:"dependencies,omitempty"`           // Dependencies module name
	Files                 []string                 `json:"files,omitempty"`                  // Files all files in the project directory
	Submodules            []*ProjectModule         `json:"submodules,omitempty"`             // Submodules submodules
}

func (module *ProjectModule) HasDependencyByTypeAndId(dependencyType string, dependencyId string) bool {
	if module.Dependencies == nil {
		return false
	}

	for _, dependency := range module.Dependencies {
		if dependency.Type == dependencyType && dependency.Id == dependencyId {
			return true
		}
	}

	return false
}

// ProjectModuleDiscovery contains info on the files used to discover the module
type ProjectModuleDiscovery struct {
	File string `json:"file"`
}

type ConfigV1Response struct {
	Debug        bool              `json:"debug,omitempty"`
	Log          map[string]string `json:"log,omitempty"`
	ProjectDir   string            `json:"project_dir,omitempty"`
	TempDir      string            `json:"temp_dir,omitempty"`
	ArtifactDir  string            `json:"artifact_dir,omitempty"`
	HostName     string            `json:"host_name,omitempty"`
	HostUserId   string            `json:"host_user_id,omitempty"`
	HostUserName string            `json:"host_user_name,omitempty"`
	HostGroupId  string            `json:"host_group_id,omitempty"`
	Config       string            `json:"config,omitempty"`
}

type ProjectExecutionContextV1Response struct {
	ProjectDir string            `json:"project-dir"`
	Config     *ConfigV1Response `json:"config"`
	Env        map[string]string `json:"env"`
	Modules    []*ProjectModule  `json:"modules"`
}

type ModuleExecutionContextV1Response struct {
	ProjectDir string            `json:"project-dir"`
	Config     *ConfigV1Response `json:"config"`
	Env        map[string]string `json:"env"`
	Module     *ProjectModule
	Deployment *DeploymentV1Response
}

type EnvironmentV1Response struct {
	Env map[string]string `json:"env"`
}

type DeploymentV1Response struct {
	DeploymentType        string            `json:"deployment_type"`
	DeploymentSpec        string            `json:"deployment_spec"`
	DeploymentEnvironment string            `json:"deployment_environment"`
	DeploymentFile        string            `json:"deployment_file"`
	Properties            map[string]string `json:"properties"`
}
