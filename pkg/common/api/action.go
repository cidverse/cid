package api

// ActionDetails holds details about the action
type ActionDetails struct {
	Stage string
	Name string
	Version string
	UsedTools []string
}

// Normalizer is a common interface to work with all normalizers
type ActionStep interface {
	GetDetails(projectDir string, env map[string]string) ActionDetails
	SetConfig(config string)
	Check(projectDir string, env map[string]string) bool
	Execute(projectDir string, env map[string]string, args []string)
}

var BuiltinActions = make(map[string]ActionStep)

// RegisterBuiltinAction registers a builtin action
func RegisterBuiltinAction(action ActionStep) {
	BuiltinActions[action.GetDetails("", nil).Name] = action
}
