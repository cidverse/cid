package provenance

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/util"

	"github.com/cidverse/cidverseutils/redact"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	slsa "github.com/in-toto/attestation/go/predicates/provenance/v1"
	attestation "github.com/in-toto/attestation/go/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	serviceGitHubActions = "github-actions"
	serviceGitLabCI      = "gitlab-ci"
)

// Step describes the plan step that produced the subject.
type Step struct {
	Name         string
	Slug         string
	Stage        string
	Action       string
	Module       string
	Environment  string
	UsesOutputOf []string
}

// Input holds everything the provenance is generated from.
type Input struct {
	NCI   nci.Spec
	Env   map[string]string
	State *state.ActionStateContext
	Step  Step
	// Now is the time the build finished, defaults to time.Now()
	Now time.Time
}

// GeneratePredicate generates the SLSA v1 provenance predicate.
func GeneratePredicate(in Input) (*slsa.Provenance, error) {
	external, err := structpb.NewStruct(externalParameters(in))
	if err != nil {
		return nil, fmt.Errorf("failed to encode external parameters: %w", err)
	}
	internal, err := structpb.NewStruct(internalParameters(in.NCI))
	if err != nil {
		return nil, fmt.Errorf("failed to encode internal parameters: %w", err)
	}

	finishedAt := in.Now
	if finishedAt.IsZero() {
		finishedAt = time.Now()
	}

	return &slsa.Provenance{
		BuildDefinition: &slsa.BuildDefinition{
			BuildType:            BuildType,
			ExternalParameters:   external,
			InternalParameters:   internal,
			ResolvedDependencies: resolvedDependencies(in),
		},
		RunDetails: &slsa.RunDetails{
			Builder: &slsa.Builder{
				Id:                  builderID(in.NCI, in.Env),
				Version:             builderVersion(),
				BuilderDependencies: builderDependencies(),
			},
			Metadata: &slsa.BuildMetadata{
				InvocationId: invocationID(in.NCI),
				StartedOn:    parseTimestamp(in.NCI.Pipeline.JobStartedAt),
				FinishedOn:   timestamppb.New(finishedAt.UTC()),
			},
		},
	}, nil
}

// Subject is the artifact a provenance statement is generated for.
type Subject struct {
	// Name is the stored file name of the artifact.
	Name string
	// SHA256 is the hex encoded SHA-256 digest of the artifact.
	SHA256 string
	// Module is the slug of the module that produced the artifact.
	Module string
}

// GenerateStatement generates an in-toto v1 statement with a SLSA v1 provenance predicate for the given subject,
// serialized as JSON.
//
// The subject name is the name the artifact was stored under. Releases prefix it with the module slug
// (`<module>_<name>`) when several modules produce the same file name, which is only known at release time - so the
// module and step are recorded as subject annotations, and verifiers match subjects by digest (in-toto treats the
// name as informational).
func GenerateStatement(subject Subject, in Input) ([]byte, error) {
	predicate, err := GeneratePredicate(in)
	if err != nil {
		return nil, err
	}
	if err = predicate.Validate(); err != nil {
		return nil, fmt.Errorf("invalid provenance predicate: %w", err)
	}

	predicateStruct, err := toStruct(predicate)
	if err != nil {
		return nil, err
	}

	statement := &attestation.Statement{
		Type: attestation.StatementTypeUri,
		Subject: []*attestation.ResourceDescriptor{
			{Name: subject.Name, Digest: map[string]string{"sha256": subject.SHA256}, Annotations: subjectAnnotations(subject, in.Step)},
		},
		PredicateType: PredicateType,
		Predicate:     predicateStruct,
	}
	if err = statement.Validate(); err != nil {
		return nil, fmt.Errorf("invalid in-toto statement: %w", err)
	}

	out, err := protojson.Marshal(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to encode in-toto statement: %w", err)
	}

	// protojson intentionally randomizes whitespace, compact it for a stable output
	var compact bytes.Buffer
	if err = json.Compact(&compact, out); err != nil {
		return nil, err
	}

	return redactStatement(compact.String())
}

// redactStatement masks registered secrets in the serialized statement. No field is expected to contain one (values are
// read from known, non-secret sources and URL credentials are stripped), this is a safety net as the provenance is
// published with the release.
func redactStatement(statement string) ([]byte, error) {
	redacted := redact.Redact(statement)
	if redacted == statement {
		return []byte(statement), nil
	}

	log.Warn().Msg("provenance contained protected values, they have been redacted")
	if !json.Valid([]byte(redacted)) {
		return nil, errors.New("provenance is invalid after redacting protected values")
	}

	return []byte(redacted), nil
}

// subjectAnnotations records where the subject comes from, to map it to its release asset name.
func subjectAnnotations(subject Subject, step Step) *structpb.Struct {
	annotations := map[string]interface{}{}
	putIfNotEmpty(annotations, "module", subject.Module)
	putIfNotEmpty(annotations, "step", step.Slug)
	if len(annotations) == 0 {
		return nil
	}

	result, _ := structpb.NewStruct(annotations)
	return result
}

// toStruct converts a proto message into a structpb.Struct, using the proto JSON field names (camelCase).
func toStruct(predicate *slsa.Provenance) (*structpb.Struct, error) {
	raw, err := protojson.Marshal(predicate)
	if err != nil {
		return nil, fmt.Errorf("failed to encode provenance predicate: %w", err)
	}

	result := &structpb.Struct{}
	if err = protojson.Unmarshal(raw, result); err != nil {
		return nil, fmt.Errorf("failed to convert provenance predicate: %w", err)
	}

	return result, nil
}

// externalParameters are the parameters under the control of the user that started the build.
func externalParameters(in Input) map[string]interface{} {
	params := map[string]interface{}{
		"source": sourceURI(in.NCI),
	}

	workflow := map[string]interface{}{}
	putIfNotEmpty(workflow, "path", in.NCI.Pipeline.ConfigFile)
	putIfNotEmpty(workflow, "ref", in.NCI.Commit.RefVCS)
	putIfNotEmpty(workflow, "repository", util.StripURLCredentials(in.NCI.Project.Url))
	putIfNotEmpty(workflow, "trigger", in.NCI.Pipeline.Trigger)
	if len(workflow) > 0 {
		params["workflow"] = workflow
	}

	step := map[string]interface{}{}
	putIfNotEmpty(step, "name", in.Step.Name)
	putIfNotEmpty(step, "slug", in.Step.Slug)
	putIfNotEmpty(step, "stage", in.Step.Stage)
	putIfNotEmpty(step, "action", in.Step.Action)
	putIfNotEmpty(step, "module", in.Step.Module)
	putIfNotEmpty(step, "environment", in.Step.Environment)
	if len(step) > 0 {
		params["step"] = step
	}

	if in.NCI.MergeRequest.Id != "" {
		mergeRequest := map[string]interface{}{"id": in.NCI.MergeRequest.Id}
		putIfNotEmpty(mergeRequest, "source_branch", in.NCI.MergeRequest.SourceBranchName)
		putIfNotEmpty(mergeRequest, "target_branch", in.NCI.MergeRequest.TargetBranchName)
		params["merge_request"] = mergeRequest
	}

	return params
}

// internalParameters are set by the build platform, they are informational and not under the control of the user.
// Only curated, non-sensitive values are recorded: the provenance is published with the release.
func internalParameters(spec nci.Spec) map[string]interface{} {
	params := map[string]interface{}{}

	platform := map[string]interface{}{}
	putIfNotEmpty(platform, "service", spec.ServiceSlug)
	putIfNotEmpty(platform, "nci_version", spec.Version)
	if len(platform) > 0 {
		params["platform"] = platform
	}

	runner := map[string]interface{}{}
	putIfNotEmpty(runner, "id", spec.Worker.Id)
	putIfNotEmpty(runner, "name", spec.Worker.Name)
	putIfNotEmpty(runner, "type", spec.Worker.Type)
	putIfNotEmpty(runner, "os", spec.Worker.OS)
	putIfNotEmpty(runner, "arch", spec.Worker.Arch)
	putIfNotEmpty(runner, "version", spec.Worker.Version)
	if len(runner) > 0 {
		params["runner"] = runner
	}

	pipeline := map[string]interface{}{}
	putIfNotEmpty(pipeline, "id", spec.Pipeline.Id)
	putIfNotEmpty(pipeline, "attempt", spec.Pipeline.Attempt)
	putIfNotEmpty(pipeline, "job_id", spec.Pipeline.JobId)
	putIfNotEmpty(pipeline, "job_name", spec.Pipeline.JobName)
	putIfNotEmpty(pipeline, "url", spec.Pipeline.Url)
	if len(pipeline) > 0 {
		params["pipeline"] = pipeline
	}

	return params
}

// resolvedDependencies lists the source revision, the actions / tools that ran and the artifacts consumed from earlier steps.
func resolvedDependencies(in Input) []*attestation.ResourceDescriptor {
	var deps []*attestation.ResourceDescriptor

	// source
	if source := sourceURI(in.NCI); source != "" {
		deps = append(deps, &attestation.ResourceDescriptor{
			Uri:    source,
			Digest: gitCommitDigest(in.NCI.Commit.Hash),
		})
	}

	if in.State == nil {
		return deps
	}

	// actions / commands
	seen := map[string]bool{}
	for _, record := range in.State.AuditLog {
		var name string
		switch record.Type {
		case "action":
			name = record.Payload["action"]
		case "command":
			name = record.Payload["binary"]
		default:
			continue
		}

		uri := record.Payload["uri"]
		digest := digestFromReference(uri)
		if !isMeaningfulURI(uri) && len(digest) == 0 {
			// builtin actions have no container reference, they are covered by the builder version
			continue
		}
		key := name + "|" + uri
		if seen[key] {
			continue
		}
		seen[key] = true

		deps = append(deps, &attestation.ResourceDescriptor{
			Name:        name,
			Uri:         uri,
			Digest:      digest,
			Annotations: referenceAnnotations(record.Payload["reference"], record.Payload["version"], digest),
		})
	}

	// input artifacts, produced by steps this step consumes the output of
	artifacts := make([]state.ActionArtifact, 0)
	for _, artifact := range in.State.Artifacts {
		if artifact.SHA256 == "" || artifact.StepSlug == in.Step.Slug || !slices.Contains(in.Step.UsesOutputOf, artifact.StepSlug) {
			continue
		}
		artifacts = append(artifacts, artifact)
	}
	slices.SortFunc(artifacts, func(a, b state.ActionArtifact) int { return strings.Compare(a.ArtifactID, b.ArtifactID) })
	for _, artifact := range artifacts {
		annotations := map[string]interface{}{"step": artifact.StepSlug, "type": artifact.Type}
		putIfNotEmpty(annotations, "module", artifact.Module)
		putIfNotEmpty(annotations, "format", artifact.Format)
		putIfNotEmpty(annotations, "format_version", artifact.FormatVersion)
		annotationStruct, _ := structpb.NewStruct(annotations)

		deps = append(deps, &attestation.ResourceDescriptor{
			Name:        artifact.Name,
			Digest:      map[string]string{"sha256": artifact.SHA256},
			Annotations: annotationStruct,
		})
	}

	return deps
}

// sourceURI returns the SPDX download location style source reference, e.g. `git+https://github.com/org/repo@refs/heads/main`.
func sourceURI(spec nci.Spec) string {
	remote := util.StripURLCredentials(spec.Repository.Remote)
	if remote == "" {
		return ""
	}

	uri := remote
	if spec.Repository.Kind != "" && !strings.HasPrefix(remote, spec.Repository.Kind+"+") {
		uri = spec.Repository.Kind + "+" + remote
	}
	if spec.Commit.RefVCS != "" {
		uri += "@" + spec.Commit.RefVCS
	}

	return uri
}

// gitCommitDigest returns the digest set of a git commit. `gitCommit` is the in-toto digest key for (SHA-1) git commit
// ids, `sha1` is added for compatibility with verifiers that predate it. The in-toto digest set has no key for SHA-256
// object ids yet, so those (and invalid values) yield no digest - the source is then only identified by its uri.
func gitCommitDigest(hash string) map[string]string {
	if len(hash) != 40 {
		return nil
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return nil
	}

	return map[string]string{"gitCommit": hash, "sha1": hash}
}

// builderID identifies the trusted build platform, following the conventions of the platforms' own provenance.
func builderID(spec nci.Spec, env map[string]string) string {
	switch spec.ServiceSlug {
	case serviceGitHubActions:
		// e.g. https://github.com/org/repo/.github/workflows/cid-main.yml@refs/heads/main
		if server, ref := env["GITHUB_SERVER_URL"], env["GITHUB_WORKFLOW_REF"]; server != "" && ref != "" {
			return strings.TrimSuffix(server, "/") + "/" + ref
		}
	case serviceGitLabCI:
		// e.g. https://gitlab.com/group/project/-/runners/12345
		if project, runner := util.StripURLCredentials(spec.Project.Url), env["CI_RUNNER_ID"]; project != "" && runner != "" {
			return strings.TrimSuffix(project, "/") + "/-/runners/" + runner
		}
	}

	return DefaultBuilderID
}

func builderVersion() map[string]string {
	version := map[string]string{"cid": constants.Version}
	if constants.CommitHash != "" && constants.CommitHash != "none" {
		version["cid-commit"] = constants.CommitHash
	}
	return version
}

// builderDependencies records the cid binary that orchestrated the build.
func builderDependencies() []*attestation.ResourceDescriptor {
	if constants.BinaryHash == "" {
		return nil
	}

	return []*attestation.ResourceDescriptor{
		{
			Name:   "cid",
			Uri:    fmt.Sprintf("https://github.com/cidverse/cid/releases/tag/v%s", strings.TrimPrefix(constants.Version, "v")),
			Digest: map[string]string{"sha256": constants.BinaryHash},
		},
	}
}

// invocationID returns a URL identifying the job run where available.
func invocationID(spec nci.Spec) string {
	if spec.Pipeline.Url != "" {
		if spec.ServiceSlug == serviceGitHubActions && spec.Pipeline.Attempt != "" {
			return strings.TrimSuffix(spec.Pipeline.Url, "/") + "/attempts/" + spec.Pipeline.Attempt
		}
		return spec.Pipeline.Url
	}
	if spec.Pipeline.Id == "" {
		return ""
	}
	if spec.Pipeline.Attempt != "" {
		return spec.Pipeline.Id + "-" + spec.Pipeline.Attempt
	}
	return spec.Pipeline.Id
}

// parseTimestamp parses a RFC3339 timestamp, unknown or invalid values yield nil (omitted) instead of the zero time.
func parseTimestamp(value string) *timestamppb.Timestamp {
	if value == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return timestamppb.New(t.UTC())
}

// referenceAnnotations records the human readable reference (e.g. the tag) and version of a resolved dependency for
// informational purposes. The uri is the authoritative, digest-pinned identifier, so the reference is only added when
// the dependency is pinned.
func referenceAnnotations(reference string, version string, digest map[string]string) *structpb.Struct {
	annotations := map[string]interface{}{}
	if reference != "" && len(digest) > 0 {
		annotations["org.opencontainers.image.ref.name"] = reference
	}
	putIfNotEmpty(annotations, "version", version)
	if len(annotations) == 0 {
		return nil
	}

	result, _ := structpb.NewStruct(annotations)
	return result
}

// isMeaningfulURI reports whether the uri identifies something beyond a bare scheme (builtin actions yield "oci://").
func isMeaningfulURI(uri string) bool {
	_, rest, found := strings.Cut(uri, "://")
	if !found {
		return uri != ""
	}
	return rest != ""
}

// digestFromReference extracts the digest from a digest-pinned reference such as
// "oci://ghcr.io/cidverse/actions/build-go@sha256:abc...". References without a digest yield an empty set.
func digestFromReference(reference string) map[string]string {
	index := strings.LastIndex(reference, "@")
	if index == -1 || index == len(reference)-1 {
		return map[string]string{}
	}

	return digestSet(reference[index+1:])
}

// digestSet converts a digest reference such as "sha256:abc..." into the algorithm to hex map used by
// in-toto resource descriptors. An empty or malformed digest yields an empty set rather than a bogus entry.
func digestSet(digest string) map[string]string {
	algorithm, value, ok := strings.Cut(digest, ":")
	if !ok || algorithm == "" || value == "" {
		return map[string]string{}
	}

	return map[string]string{algorithm: value}
}

func putIfNotEmpty(m map[string]interface{}, key string, value string) {
	if value != "" {
		m[key] = value
	}
}
