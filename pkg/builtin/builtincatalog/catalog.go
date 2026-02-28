package builtincatalog

import (
	_ "embed"
	"log/slog"
	"os"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction"
	"github.com/cidverse/cid/pkg/builtin/builtinworkflow"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/lib/files"
)

//go:embed files/cid-index.json
var indexJSON []byte

// InternalCatalog returns a virtual catalog with all built-in actions and workflows
func InternalCatalog() catalog.Config {
	var actions []catalog.Action
	for _, action := range builtinaction.GetActions(nil) { // TODO: dummy sdk client for metadata collection
		am := action.Metadata()
		catalogActionMetadata := convertActionMetadata(am)

		act := catalog.Action{
			Repository: "builtin",
			URI:        "builtin://actions/" + am.Name,
			Type:       catalog.ActionTypeBuiltIn,
			Container:  catalog.ContainerAction{},
			Version:    constants.Version,
			Metadata:   catalogActionMetadata,
		}

		actions = append(actions, act)
	}

	embeddedCatalog, err := files.ReadJson[catalog.Config](indexJSON)
	if err != nil {
		slog.With("err", err).Error("failed to read embedded json index")
		os.Exit(1)
	}

	return catalog.Config{
		Actions:             actions,
		Workflows:           builtinworkflow.GetWorkflows(),
		ExecutableDiscovery: nil,
		Executables:         embeddedCatalog.Executables,
	}
}

func convertActionMetadata(actionMetadata cidsdk.ActionMetadata) catalog.ActionMetadata {
	var workflowRules []catalog.WorkflowRule
	for _, rule := range actionMetadata.Rules {
		workflowRules = append(workflowRules, catalog.WorkflowRule{
			Type:       catalog.WorkflowExpressionType(rule.Type),
			Expression: rule.Expression,
		})
	}

	var accessEnvironment []catalog.ActionAccessEnv
	for _, env := range actionMetadata.Access.Environment {
		accessEnvironment = append(accessEnvironment, catalog.ActionAccessEnv{
			Name:        env.Name,
			Description: env.Description,
			Required:    env.Required,
			Secret:      env.Secret,
		})
	}

	var accessExecutable []catalog.ActionAccessExecutable
	for _, exec := range actionMetadata.Access.Executables {
		accessExecutable = append(accessExecutable, catalog.ActionAccessExecutable{
			Name:       exec.Name,
			Constraint: exec.Constraint,
		})
	}

	var accessNetwork []catalog.ActionAccessNetwork
	for _, net := range actionMetadata.Access.Network {
		accessNetwork = append(accessNetwork, catalog.ActionAccessNetwork{
			Host: net.Host,
		})
	}

	var inputArtifacts []catalog.ActionArtifactType
	for _, artifact := range actionMetadata.Input.Artifacts {
		inputArtifacts = append(inputArtifacts, catalog.ActionArtifactType{
			Type:   artifact.Type,
			Format: artifact.Format,
		})
	}

	var outputArtifacts []catalog.ActionArtifactType
	for _, artifact := range actionMetadata.Output.Artifacts {
		outputArtifacts = append(outputArtifacts, catalog.ActionArtifactType{
			Type:   artifact.Type,
			Format: artifact.Format,
		})
	}

	return catalog.ActionMetadata{
		Name:          actionMetadata.Name,
		Description:   actionMetadata.Description,
		Documentation: actionMetadata.Documentation,
		Category:      actionMetadata.Category,
		Scope:         catalog.ActionScope(actionMetadata.Scope),
		Links:         actionMetadata.Links,
		Rules:         workflowRules,
		RunIfChanged:  actionMetadata.RunIfChanged,
		Access: catalog.ActionAccess{
			Environment: accessEnvironment,
			Executables: accessExecutable,
			Network:     accessNetwork,
		},
		Input: catalog.ActionInput{
			Artifacts: inputArtifacts,
		},
		Output: catalog.ActionOutput{
			Artifacts: outputArtifacts,
		},
	}
}
