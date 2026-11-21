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
)
