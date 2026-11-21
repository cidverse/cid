package provenance

// FileTypes ist a list of file types that we will automatically generate a provenance for when seen
var FileTypes = []string{
	"binary",
}

const (
	// ArtifactType is the artifact type of generated provenance statements.
	ArtifactType = "attestation"
	// ArtifactFormat is the artifact format of generated provenance statements.
	ArtifactFormat = "provenance"
	// FileExtension is appended to the artifact name to name its provenance statement (`<artifact>.provenance.json`).
	FileExtension = ".provenance.json"

	// PredicateType is the SLSA provenance predicate type (SLSA v1.0 / v1.1 share this URI).
	PredicateType = "https://slsa.dev/provenance/v1"
	// BuildType identifies the cid build type, which defines how to interpret external / internal parameters
	// (documented in docs/provenance.md). Not versioned yet, see ROADMAP.md.
	BuildType = "https://github.com/cidverse/cid"
	// DefaultBuilderID is used when the CI platform does not provide a more specific builder identity (e.g. local runs).
	DefaultBuilderID = "https://github.com/cidverse/cid"
)
