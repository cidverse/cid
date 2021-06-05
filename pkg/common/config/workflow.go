package config

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/rs/zerolog/log"
)

type WorkflowStage struct {
	Name string
	Rules []WorkflowRule
	Actions []WorkflowAction
}

type WorkflowAction struct {
	Name string `required:"true"`
	Type string `default:"builtin"`
	Config interface{} `yaml:"config,omitempty"`
}

type WorkflowRule struct {
	Expression string
}

// FindWorkflowStages finds all relevant stages for the current context (branch, tag, ...)
func FindWorkflowStages(projectDir string, env map[string]string) []WorkflowStage {
	// cel expr environment
	celConfig, celConfigErr := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("NCI_COMMIT_REF_PATH", decls.String),
		),
	)
	if celConfigErr != nil {
		log.Fatal().Err(celConfigErr).Msg("failed to initialize CEL")
	}

	inputData := map[string]interface{}{
		"NCI_COMMIT_REF_PATH": env["NCI_COMMIT_REF_PATH"],
	}

	var activeStages []WorkflowStage
	for _, stage := range Config.Stages {
		if len(stage.Rules) > 0 {
			// evaluate rules
			for _, rule := range stage.Rules {
				if len(rule.Expression) > 0 {
					log.Debug().Str("stage", stage.Name).Str("expression", rule.Expression).Msg("checking expression for stage rule")

					// prepare program for evaluation
					ast, issues := celConfig.Compile(rule.Expression)
					if issues != nil && issues.Err() != nil {
						log.Fatal().Err(issues.Err()).Msg("stage rule type error: " + issues.Err().Error())
					}
					prg, prgErr := celConfig.Program(ast)
					if prgErr != nil {
						log.Fatal().Err(prgErr).Msg("program construction error")
					}

					// evaluate
					execOut, _, execErr := prg.Eval(inputData)
					if execErr != nil {
						log.Warn().Err(execErr).Msg("failed to evaluate filter rule")
					}

					// check result
					if execOut.Type() == types.BoolType {
						if fmt.Sprintf("%+v", execOut) == "true" {
							activeStages = append(activeStages, stage)
							break
						}
					} else {
						log.Warn().Str("stage", stage.Name).Str("expression", rule.Expression).Msg("ignoring stage rule expression, does not return a boolean")
					}
				} else {
					log.Warn().Str("stage", stage.Name).Str("rule", fmt.Sprintf("%+v", rule)).Msg("stage rule can't be evaluated")
				}
			}

		} else {
			activeStages = append(activeStages, stage)
		}
	}

	return activeStages
}
