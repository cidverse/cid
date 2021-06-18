package analyzerapi

var Analyzers []Analyzer

// Analyzer is the interface that needs to be implemented by all analyzers
type Analyzer interface {
	// Analyze will retrieve information about the project
	Analyze(projectDir string) []*ProjectModule
}

// ProjectModule
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
	BuildSystemSyntax *ProjectBuildSystemSyntax

	// Language of the project
	Language map[ProjectLanguage]*string

	// Dependencies
	Dependencies []ProjectDependency

	// Submodules contains information about submodules
	Submodules []*ProjectModule
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
	BuildSystemGradle ProjectBuildSystem = "gradle"
	BuildSystemGoMod  ProjectBuildSystem = "gomod"
	BuildSystemNpm    ProjectBuildSystem = "npm"
)

type ProjectBuildSystemSyntax string

const (
	GradleGroovyDSL ProjectBuildSystemSyntax = "groovy"
	GradleKotlinDSL ProjectBuildSystemSyntax = "kotlin"
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
