package provenance

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	slsa "github.com/in-toto/attestation/go/predicates/provenance/v1"
	attestation "github.com/in-toto/attestation/go/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// ReleaseSubject is a release file covered by the release statement.
type ReleaseSubject struct {
	// Name is the release asset name.
	Name string
	// SHA256 is the hex encoded SHA-256 digest of the file.
	SHA256 string
	// Module is the slug of the module that produced the file.
	Module string
	// Step is the slug of the plan step that produced the file.
	Step string
	// Type and Format are the artifact type / format of the file (e.g. `binary`, `sbom` / `cyclonedx-json`).
	Type   string
	Format string
}

// GenerateReleaseStatement generates a single in-toto v1 statement with a SLSA v1 provenance predicate that covers every
// release file as a subject, serialized as JSON. Signed as DSSE envelope it lets users verify any release file with
// one bundle (`cosign verify-blob-attestation`, `gh attestation verify`), matched by digest.
//
// The predicate describes the pipeline run (source, workflow, builder, invocation) as seen by the job generating it, in.
// in.Step and in.State are ignored, per-step details are recorded as subject annotations instead. The resolved
// dependencies of buildStatements (the per-artifact statements of GenerateStatement, written by the build steps) are
// merged in, so the actions / tools / inputs of every build step are listed once. Job specific internal parameters
// (runner, job id / name) are dropped as they differ between the jobs of the run.
func GenerateReleaseStatement(subjects []ReleaseSubject, in Input, buildStatements [][]byte) ([]byte, error) {
	if len(subjects) == 0 {
		return nil, fmt.Errorf("release statement needs at least one subject")
	}

	predicate, err := GeneratePredicate(Input{NCI: in.NCI, Env: in.Env, Now: in.Now})
	if err != nil {
		return nil, err
	}
	dropJobParameters(predicate.GetBuildDefinition().GetInternalParameters())

	// merge the resolved dependencies of the build steps
	seen := map[string]bool{}
	for _, dep := range predicate.GetBuildDefinition().GetResolvedDependencies() {
		seen[dependencyKey(dep)] = true
	}
	for i, raw := range buildStatements {
		buildPredicate, parseErr := parseProvenanceStatement(raw)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse build provenance %d: %w", i, parseErr)
		}
		for _, dep := range buildPredicate.GetBuildDefinition().GetResolvedDependencies() {
			key := dependencyKey(dep)
			if seen[key] {
				continue
			}
			seen[key] = true
			predicate.BuildDefinition.ResolvedDependencies = append(predicate.BuildDefinition.ResolvedDependencies, dep)
		}
		// the run started with the earliest build job
		if started := buildPredicate.GetRunDetails().GetMetadata().GetStartedOn(); started != nil {
			metadata := predicate.GetRunDetails().GetMetadata()
			if metadata.GetStartedOn() == nil || started.AsTime().Before(metadata.GetStartedOn().AsTime()) {
				metadata.StartedOn = started
			}
		}
	}

	if err = predicate.Validate(); err != nil {
		return nil, fmt.Errorf("invalid provenance predicate: %w", err)
	}
	predicateStruct, err := toStruct(predicate)
	if err != nil {
		return nil, err
	}

	sorted := slices.Clone(subjects)
	slices.SortFunc(sorted, func(a, b ReleaseSubject) int { return strings.Compare(a.Name, b.Name) })
	descriptors := make([]*attestation.ResourceDescriptor, 0, len(sorted))
	for _, subject := range sorted {
		if subject.Name == "" || subject.SHA256 == "" {
			return nil, fmt.Errorf("release subject %q has no name or digest", subject.Name)
		}
		descriptors = append(descriptors, &attestation.ResourceDescriptor{
			Name:        subject.Name,
			Digest:      map[string]string{"sha256": subject.SHA256},
			Annotations: releaseSubjectAnnotations(subject),
		})
	}

	return marshalStatement(&attestation.Statement{
		Type:          attestation.StatementTypeUri,
		Subject:       descriptors,
		PredicateType: PredicateType,
		Predicate:     predicateStruct,
	})
}

// parseProvenanceStatement parses an in-toto statement and returns its SLSA v1 provenance predicate.
func parseProvenanceStatement(raw []byte) (*slsa.Provenance, error) {
	statement := &attestation.Statement{}
	if err := protojson.Unmarshal(raw, statement); err != nil {
		return nil, err
	}
	if statement.GetPredicateType() != PredicateType {
		return nil, fmt.Errorf("unexpected predicate type %q", statement.GetPredicateType())
	}

	predicateJSON, err := protojson.Marshal(statement.GetPredicate())
	if err != nil {
		return nil, err
	}
	predicate := &slsa.Provenance{}
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(predicateJSON, predicate); err != nil {
		return nil, err
	}
	return predicate, nil
}

// dropJobParameters removes the internal parameters that describe a single job of the run.
func dropJobParameters(params *structpb.Struct) {
	if params == nil {
		return
	}
	delete(params.Fields, "runner")
	if pipeline := params.Fields["pipeline"].GetStructValue(); pipeline != nil {
		delete(pipeline.Fields, "job_id")
		delete(pipeline.Fields, "job_name")
	}
}

// dependencyKey identifies a resolved dependency for de-duplication.
func dependencyKey(dep *attestation.ResourceDescriptor) string {
	digests := make([]string, 0, len(dep.GetDigest()))
	for _, algorithm := range slices.Sorted(maps.Keys(dep.GetDigest())) {
		digests = append(digests, algorithm+":"+dep.GetDigest()[algorithm])
	}
	return dep.GetName() + "|" + dep.GetUri() + "|" + strings.Join(digests, ",")
}

func releaseSubjectAnnotations(subject ReleaseSubject) *structpb.Struct {
	annotations := map[string]interface{}{}
	putIfNotEmpty(annotations, "module", subject.Module)
	putIfNotEmpty(annotations, "step", subject.Step)
	putIfNotEmpty(annotations, "type", subject.Type)
	putIfNotEmpty(annotations, "format", subject.Format)
	if len(annotations) == 0 {
		return nil
	}

	result, _ := structpb.NewStruct(annotations)
	return result
}
