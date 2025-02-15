package catalog

func ProcessCatalog(catalog *Config) *Config {
	result := Config{
		Actions:   make([]Action, len(catalog.Actions)),
		Workflows: make([]Workflow, len(catalog.Workflows)),
	}

	// actions
	for _, sourceAction := range catalog.Actions { //nolint:gocritic
		sourceAction.Repository = ""
		result.Actions = append(result.Actions, sourceAction)
	}

	// workflows
	for _, sourceWorkflow := range catalog.Workflows { //nolint:gocritic
		sourceWorkflow.Repository = ""
		result.Workflows = append(result.Workflows, sourceWorkflow)
	}

	return &result
}
