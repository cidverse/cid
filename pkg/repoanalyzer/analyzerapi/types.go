package analyzerapi

var Analyzers []Analyzer

// Analyzer is the interface that needs to be implemented by all analyzers
type Analyzer interface {
	// GetName returns the name of the analyzer
	GetName() string

	// Analyze will retrieve information about the project
	Analyze(ctx AnalyzerContext) []*ProjectModule
}

// ProjectModule contains information about project modules
type ProjectModule struct {
	// RootDirectory stores the project root directory
	RootDirectory string

	// Directory stores the module root directory
	Directory string

	// Discovery stores information on how this module was discovered
	Discovery string

	// Name stores the module name
	Name string

	// Slug contains a url/folder name compatible name of the module
	Slug string

	// BuildSystem used in this project
	BuildSystem ProjectBuildSystem

	// BuildSystemSyntax used in this project
	BuildSystemSyntax ProjectBuildSystemSyntax

	// Language of the project
	Language map[ProjectLanguage]*string

	// Dependencies
	Dependencies []ProjectDependency

	// Submodules contains information about submodules
	Submodules []*ProjectModule

	// Files holds all project files
	Files []string

	// FilesByExtension contains all files by extension
	FilesByExtension map[string][]string
}

type ProjectLanguage string

const (
	LanguageGolang     ProjectLanguage = "go"
	LanguageJava       ProjectLanguage = "java"
	LanguageJavascript ProjectLanguage = "javascript"
	LanguageTypescript ProjectLanguage = "typescript"
)

type ProjectBuildSystem string

const (
	BuildSystemGradle    ProjectBuildSystem = "gradle"
	BuildSystemMaven     ProjectBuildSystem = "maven"
	BuildSystemGoMod     ProjectBuildSystem = "gomod"
	BuildSystemNpm       ProjectBuildSystem = "npm"
	BuildSystemHugo      ProjectBuildSystem = "hugo"
	BuildSystemHelm      ProjectBuildSystem = "helm"
	BuildSystemContainer ProjectBuildSystem = "container"
)

type ProjectBuildSystemSyntax string

const (
	BuildSystemSyntaxDefault ProjectBuildSystemSyntax = "default"
	GradleGroovyDSL          ProjectBuildSystemSyntax = "groovy"
	GradleKotlinDSL          ProjectBuildSystemSyntax = "kotlin"
	ContainerDockerfile      ProjectBuildSystemSyntax = "dockerfile"
	ContainerBuildahScript   ProjectBuildSystemSyntax = "buildah-script"
)

// ProjectDependency contains dependency information
type ProjectDependency struct {
	// Type is the dep kind
	Type string

	// Id is the identifier
	Id string

	// Version is the dep version
	Version string
}

// AnalyzerContext holds the context to analyze projects
type AnalyzerContext struct {
	// ProjectDir holds the project directory
	ProjectDir string

	// Files holds all project files
	Files []string

	// FilesByExtension contains all files by extension
	FilesByExtension map[string][]string

	// FilesWithoutExtension contains all files without an extension
	FilesWithoutExtension []string
}
