package builtincatalog

import (
	_ "embed"
	"log/slog"
	"os"

	"github.com/cidverse/cid/pkg/builtin/builtinaction"
	"github.com/cidverse/cid/pkg/builtin/builtinworkflow"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/actionsdk"
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

func convertActionMetadata(actionMetadata actionsdk.ActionMetadata) catalog.ActionMetadata {
	var workflowRules []catalog.WorkflowRule
	for _, rule := range actionMetadata.Rules {
		workflowRules = append(workflowRules, catalog.WorkflowRule{
			Type:       catalog.WorkflowExpressionType(rule.Type),
			Expression: rule.Expression,
		})
	}

	return catalog.ActionMetadata{
		Name:          actionMetadata.Name,
		Description:   actionMetadata.Description,
		Documentation: actionMetadata.Documentation,
		Category:      actionMetadata.Category,
		Scope:         actionMetadata.Scope,
		Links:         actionMetadata.Links,
		Rules:         workflowRules,
		RunIfChanged:  actionMetadata.RunIfChanged,
		Access:        actionMetadata.Access,
		Input:         actionMetadata.Input,
		Output:        actionMetadata.Output,
	}
}
