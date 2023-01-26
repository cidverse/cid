package registry

// Config is a registry configuration with placeholders
type Config struct {
	// Actions
	Actions []Action `yaml:"actions,omitempty"`

	// ContainerImages
	ContainerImages []ContainerImage `yaml:"images,omitempty"`

	// Workflows
	Workflows []Workflow `yaml:"workflows,omitempty"`
}

// FindWorkflow finds a workflow by name
func (r *Config) FindWorkflow(name string) *Workflow {
	for _, w := range r.Workflows {
		if w.Name == name {
			return &w
		}
	}

	return nil
}

// FindAction finds an action by id
func (r *Config) FindAction(name string) *Action {
	// exact match
	for i := range r.Actions {
		a := r.Actions[i]
		if a.Repository+"/"+a.Name == name {
			return &a
		}
	}

	return nil
}
