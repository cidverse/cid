package provenance

import (
	"fmt"
	"time"

	"github.com/cidverse/cid/pkg/core/state"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1.0"
)

var WorkflowSource string
var Workflow string

func GenerateProvenance(env map[string]string, state *state.ActionStateContext) v1.ProvenancePredicate {
	startedAt, _ := time.Parse(time.RFC3339, env["NCI_PIPELINE_JOB_STARTED_AT"])
	prov := v1.ProvenancePredicate{}

	// builder
	var resolvedDependencies []v1.ArtifactReference
	resolvedDependencies = append(resolvedDependencies, v1.ArtifactReference{
		URI: fmt.Sprintf("%s+%s@%s", env["NCI_REPOSITORY_KIND"], env["NCI_REPOSITORY_REMOTE"], env["NCI_COMMIT_REF_NAME"]),
		Digest: common.DigestSet{
			"sha1": env["NCI_COMMIT_SHA"],
		},
	})

	resolvedDependencies = append(resolvedDependencies, v1.ArtifactReference{
		URI:              fmt.Sprintf("%s:%s", env["NCI_WORKER_TYPE"], env["NCI_WORKER_OS"]),
		Digest:           nil,
		LocalName:        "",
		DownloadLocation: "",
		MediaType:        "",
	})

	prov.BuildDefinition = v1.ProvenanceBuildDefinition{
		BuildType: fmt.Sprintf("https://github.com/cidverse/cid@%s", "0.0.0"),
		ExternalParameters: map[string]string{
			"cid-workflow-source": WorkflowSource,
			"cid-workflow":        Workflow,
			"source":              fmt.Sprintf("%s+%s@%s", env["NCI_REPOSITORY_KIND"], env["NCI_REPOSITORY_REMOTE"], env["NCI_COMMIT_REF_NAME"]),
		},
		SystemParameters: map[string]string{
			"RUNNER": fmt.Sprintf("%s:%s", env["NCI_WORKER_TYPE"], env["NCI_WORKER_OS"]),
		},
		ResolvedDependencies: resolvedDependencies,
	}

	// run details
	prov.RunDetails = v1.ProvenanaceRunDetails{
		Builder: v1.Builder{
			ID:                  fmt.Sprintf("https://github.com/cidverse/cid@%s", "0.0.0"),
			Version:             nil,
			BuilderDependencies: nil,
		},
		BuildMetadata: v1.BuildMetadata{
			InvocationID: "",
			StartedOn:    &startedAt,
			FinishedOn:   nil,
		},
		Byproducts: nil,
	}

	return prov
}
