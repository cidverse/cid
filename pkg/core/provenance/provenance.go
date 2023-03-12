package provenance

import (
	"fmt"
	"time"

	"github.com/PhilippHeuer/in-toto-golang/in_toto/slsa_provenance/v1.0"
	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/normalizeci/pkg/ncispec"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
)

var WorkflowSource string
var Workflow string

func GeneratePredicate(env map[string]string, state *state.ActionStateContext) v1.ProvenancePredicate {
	nci := ncispec.OfMap(env)

	startedAt, _ := time.Parse(time.RFC3339, nci.PipelineJobStartedAt)
	startedAt = startedAt.UTC()
	finishedAt := time.Now().UTC()
	prov := v1.ProvenancePredicate{}

	// builder
	var resolvedDependencies []v1.ArtifactReference
	resolvedDependencies = append(resolvedDependencies, v1.ArtifactReference{
		URI: fmt.Sprintf("%s+%s@%s", nci.RepositoryKind, nci.RepositoryRemote, nci.CommitRefType),
		Digest: common.DigestSet{
			"sha1": nci.CommitSha,
		},
	})
	resolvedDependencies = append(resolvedDependencies, v1.ArtifactReference{
		URI:              fmt.Sprintf("%s:%s", nci.WorkerType, nci.WorkerOS),
		Digest:           nil,
		LocalName:        "",
		DownloadLocation: "",
		MediaType:        "",
	})

	var systemParameters = map[string]string{
		"RUNNER": fmt.Sprintf("%s:%s", nci.WorkerType, nci.WorkerOS),
	}
	for k, v := range env {
		systemParameters[protectoutput.RedactProtectedPhrases(k)] = protectoutput.RedactProtectedPhrases(v)
	}
	prov.BuildDefinition = v1.ProvenanceBuildDefinition{
		BuildType: fmt.Sprintf("https://github.com/cidverse/cid@%s", "0.0.0"),
		ExternalParameters: map[string]string{
			"cid-workflow-source": WorkflowSource,
			"cid-workflow":        Workflow,
			"source":              fmt.Sprintf("%s+%s@%s", nci.RepositoryKind, nci.RepositoryRemote, nci.CommitRefName),
		},
		SystemParameters:     systemParameters,
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
			InvocationID: fmt.Sprintf("%s-%s", nci.PipelineId, nci.PipelineAttempt),
			StartedOn:    &startedAt,
			FinishedOn:   &finishedAt,
		},
		Byproducts: nil,
	}

	return prov
}

// GenerateInTotoPredicate generates an in-toto statement with a SLSA-Predicate
func GenerateInTotoPredicate(fileName string, hash string, env map[string]string, state *state.ActionStateContext) intoto.Statement {
	return intoto.Statement{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: v1.PredicateSLSAProvenance,
			Subject: []intoto.Subject{
				{
					Name:   fileName,
					Digest: common.DigestSet{"hash": hash},
				},
			},
		},
		Predicate: GeneratePredicate(env, state),
	}
}
