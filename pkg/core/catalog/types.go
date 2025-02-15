package catalog

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

var workflowRegexp = regexp.MustCompile(`(?P<repo>\w+)/?(?P<workflow>\w+)@?(?P<version>[\w.]+)?`)

// Config is a registry configuration with placeholders
type Config struct {
	// Actions
	Actions []Action `yaml:"actions,omitempty"`

	// Workflows
	Workflows []Workflow `yaml:"workflows,omitempty"`
}

// FindWorkflow finds a workflow by name
func (r *Config) FindWorkflow(id string) *Workflow {
	for _, w := range r.Workflows {
		if isMatchingWorkflow(id, &w) {
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
		if a.Repository+"/"+a.Metadata.Name == name {
			return &a
		}
	}

	return nil
}

func isMatchingWorkflow(id string, workflow *Workflow) bool {
	// parse id
	if !strings.Contains(id, "/") {
		id = "cid/" + id
	}
	match := workflowRegexp.FindStringSubmatch(id)
	if match == nil {
		log.Fatal().Msg("invalid workflow name, please use the following format <repository>/<workflow>@<workflowVersion>")
	}
	repo := match[1]
	name := match[2]
	version := match[3]

	if workflow.Repository == repo && workflow.Name == name {
		if len(version) > 0 && workflow.Version == version {
			return true
		} else if version == "" {
			return true
		}
	}

	return false
}
