package provenance

import (
	"fmt"

	"github.com/cidverse/cid/pkg/cmd"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
)

func Generate(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, action *catalog.WorkflowAction) slsa.ProvenancePredicate {
	prov := slsa.ProvenancePredicate{}

	// builder
	prov.BuildType = fmt.Sprintf("https://github.com/cidverse/cid@%s", cmd.Version)
	prov.Builder = common.ProvenanceBuilder{
		ID: fmt.Sprintf("https://github.com/cidverse/cid@%s", cmd.Version),
	}

	// invocation
	prov.Invocation = slsa.ProvenanceInvocation{
		// Non user-controllable environment vars needed to reproduce the build.
		Environment: map[string]interface{}{
			"arch": "amd64",
			"os":   "ubuntu",
		},
		// Parameters coming from the trigger event.
		Parameters: map[string]string{},
	}

	// metadata
	prov.Metadata = &slsa.ProvenanceMetadata{
		BuildStartedOn: nil,
		Completeness: slsa.ProvenanceComplete{
			Environment: true,
		},
	}

	// materials
	prov.Materials = []common.ProvenanceMaterial{}

	return prov
}
