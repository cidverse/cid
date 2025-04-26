package provenance

import (
	"fmt"
	"github.com/cidverse/cid/internal/state"
	"time"

	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/normalizeci/pkg/envstruct"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v1"
)

var WorkflowSource string
var Workflow string

func GeneratePredicate(env map[string]string, state *state.ActionStateContext) v1.ProvenancePredicate {
	var nci nci.Spec
	envstruct.EnvMapToStruct(&nci, env)

	startedAt, _ := time.Parse(time.RFC3339, nci.Pipeline.JobStartedAt)
	startedAt = startedAt.UTC()
	finishedAt := time.Now().UTC()
	prov := v1.ProvenancePredicate{}

	// builder
	var resolvedDependencies []v1.ResourceDescriptor
	resolvedDependencies = append(resolvedDependencies, v1.ResourceDescriptor{
		URI: fmt.Sprintf("%s+%s@%s", nci.Repository.Kind, nci.Repository.Remote, nci.Commit.RefType),
		Digest: common.DigestSet{
			"sha1": nci.Commit.Hash,
		},
	})
	resolvedDependencies = append(resolvedDependencies, v1.ResourceDescriptor{
		URI:  fmt.Sprintf("%s:%s:%s", nci.Worker.Type, nci.Worker.OS, nci.Worker.Version),
		Name: fmt.Sprintf("%s:%s", nci.Worker.Type, nci.Worker.OS),
	})

	for _, record := range state.AuditLog {
		if record.Type == "action" {
			resolvedDependencies = append(resolvedDependencies, v1.ResourceDescriptor{
				URI: record.Payload["uri"],
				Digest: map[string]string{
					"sha1": record.Payload["digest"],
				},
				Name:      record.Payload["action"],
				MediaType: "application/vnd.oci.image.index.v1+json",
			})
		} else if record.Type == "command" {
			resolvedDependencies = append(resolvedDependencies, v1.ResourceDescriptor{
				URI: record.Payload["uri"],
				Digest: map[string]string{
					"sha1": record.Payload["digest"],
				},
				Name:      record.Payload["binary"],
				MediaType: "application/vnd.oci.image.index.v1+json",
			})
		}
	}

	var systemParameters = map[string]string{
		"RUNNER": fmt.Sprintf("%s:%s", nci.Worker.Type, nci.Worker.OS),
	}
	for k, v := range env {
		systemParameters[redact.Redact(k)] = redact.Redact(v)
	}
	prov.BuildDefinition = v1.ProvenanceBuildDefinition{
		BuildType: fmt.Sprintf("https://github.com/cidverse/cid@%s", "0.0.0"),
		ExternalParameters: map[string]string{
			"cid-workflow-source": WorkflowSource,
			"cid-workflow":        Workflow,
			"source":              fmt.Sprintf("%s+%s@%s", nci.Repository.Kind, nci.Repository.Remote, nci.Commit.RefName),
		},
		InternalParameters:   systemParameters,
		ResolvedDependencies: resolvedDependencies,
	}

	// run details
	prov.RunDetails = v1.ProvenanceRunDetails{
		Builder: v1.Builder{
			ID:                  fmt.Sprintf("https://github.com/cidverse/cid@%s", "0.0.0"),
			Version:             map[string]string{},
			BuilderDependencies: nil,
		},
		BuildMetadata: v1.BuildMetadata{
			InvocationID: fmt.Sprintf("%s-%s", nci.Pipeline.Id, nci.Pipeline.Attempt),
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
