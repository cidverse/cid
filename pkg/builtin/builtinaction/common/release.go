package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/provenance"
)

const (
	// ArtifactTypeChecksum is the artifact type of checksum files (per-file `<asset>.sha256` and the aggregated ChecksumsFileName).
	ArtifactTypeChecksum = "checksum"
	// ArtifactFormatSHA256 is the artifact format of checksum files, using the `sha256sum` line format.
	ArtifactFormatSHA256 = "sha256"
	// ChecksumsFileName is the aggregated checksum file listing every release binary, stored in the root module.
	ChecksumsFileName = "SHA256SUMS"
	// ChecksumsModule is the module the aggregated checksum file and its signature are stored in.
	ChecksumsModule = "root"
	// ReleaseAttestationFileName is the Sigstore bundle of the signed release provenance, covering every release file.
	ReleaseAttestationFileName = "release.intoto.sigstore.json"
	// ReleaseAttestationModule is the module the release attestation is stored in.
	ReleaseAttestationModule = "root"
)

// ReleaseArtifactQuery selects every artifact published as a release asset.
const ReleaseArtifactQuery = `(artifact_type == "binary" || artifact_type == "signature" || artifact_type == "checksum" || artifact_type == "sbom" || (artifact_type == "attestation" && format == "provenance"))`

// SignableArtifactQuery selects the artifacts to sign: binaries, their provenance statements and SBOMs.
const SignableArtifactQuery = `(artifact_type == "binary" || artifact_type == "sbom" || (artifact_type == "attestation" && format == "provenance"))`

// IsReleaseAttestation reports whether the artifact is the signed release attestation (ReleaseAttestationFileName).
func IsReleaseAttestation(artifact *actionsdk.Artifact) bool {
	return artifact.Type == "signature" && artifact.Format == "sigstore" && artifact.Name == ReleaseAttestationFileName
}

// ReleaseAssets returns the artifacts to publish as release assets, from the artifacts selected by ReleaseArtifactQuery.
// When the release attestation is present, the per-binary provenance statements are left out: the attestation lists
// every release file and merges their resolved dependencies, so they would only duplicate it unsigned.
func ReleaseAssets(artifacts []*actionsdk.Artifact) []*actionsdk.Artifact {
	if !slices.ContainsFunc(artifacts, IsReleaseAttestation) {
		return artifacts
	}
	return slices.DeleteFunc(slices.Clone(artifacts), IsProvenance)
}

// BinaryArtifacts returns the binary artifacts of the given artifacts.
func BinaryArtifacts(artifacts []*actionsdk.Artifact) []*actionsdk.Artifact {
	var binaries []*actionsdk.Artifact
	for _, artifact := range artifacts {
		if artifact.Type == "binary" {
			binaries = append(binaries, artifact)
		}
	}
	return binaries
}

// ReleaseAssetNames returns the release asset name for each artifact, keyed by ArtifactID.
//
// The asset name is the artifact name, unless several modules produce an artifact with the same name - those are
// prefixed with the module slug (`<module>_<name>`) so they don't overwrite each other in a release. Callers pass the
// same artifact set (all release binaries) so every action derives the same names.
func ReleaseAssetNames(artifacts []*actionsdk.Artifact) map[string]string {
	modulesByName := make(map[string]map[string]struct{})
	for _, a := range artifacts {
		if modulesByName[a.Name] == nil {
			modulesByName[a.Name] = make(map[string]struct{})
		}
		modulesByName[a.Name][a.Module] = struct{}{}
	}

	names := make(map[string]string, len(artifacts))
	for _, a := range artifacts {
		if len(modulesByName[a.Name]) > 1 {
			names[a.ArtifactID] = fmt.Sprintf("%s_%s", a.Module, a.Name)
		} else {
			names[a.ArtifactID] = a.Name
		}
	}
	return names
}

// IsProvenance reports whether the artifact is a generated provenance statement (`<binary>.provenance.json`).
func IsProvenance(artifact *actionsdk.Artifact) bool {
	return artifact.Type == provenance.ArtifactType && artifact.Format == provenance.ArtifactFormat
}

// ReleaseFileNames returns the release asset name for each given artifact, keyed by ArtifactID.
//
// Binaries are named by ReleaseAssetNames, provenance statements follow the binary they describe
// (`<binary asset name>.provenance.json`), so a module prefix applied to the binary is applied to its provenance too.
// SBOMs get the same collision handling as binaries (`<module>_<name>` when several modules store the same name, e.g.
// `maven-target-bom.cdx.json`), callers must pass every SBOM so all actions derive the same names.
// Everything else (signatures, checksums) is already named after the asset and keeps its name.
func ReleaseFileNames(artifacts []*actionsdk.Artifact) map[string]string {
	binaries := BinaryArtifacts(artifacts)
	binaryNames := ReleaseAssetNames(binaries)
	var sboms []*actionsdk.Artifact
	for _, a := range artifacts {
		if a.Type == SBOMArtifactType {
			sboms = append(sboms, a)
		}
	}
	sbomNames := ReleaseAssetNames(sboms)
	binaryByModuleAndName := make(map[string]*actionsdk.Artifact, len(binaries))
	for _, b := range binaries {
		binaryByModuleAndName[b.Module+"|"+b.Name] = b
	}

	names := make(map[string]string, len(artifacts))
	for _, a := range artifacts {
		switch {
		case a.Type == "binary":
			names[a.ArtifactID] = binaryNames[a.ArtifactID]
		case a.Type == SBOMArtifactType:
			names[a.ArtifactID] = sbomNames[a.ArtifactID]
		case IsProvenance(a):
			names[a.ArtifactID] = a.Name
			if b, ok := binaryByModuleAndName[a.Module+"|"+strings.TrimSuffix(a.Name, provenance.FileExtension)]; ok {
				names[a.ArtifactID] = binaryNames[b.ArtifactID] + provenance.FileExtension
			}
		default:
			names[a.ArtifactID] = a.Name
		}
	}
	return names
}

// SHA256File returns the hex encoded SHA-256 digest of file.
func SHA256File(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", file, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", file, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ChecksumLine formats a `sha256sum` compatible line ("<hex>  <name>\n"), verifiable with `sha256sum -c`.
func ChecksumLine(digest string, name string) string {
	return fmt.Sprintf("%s  %s\n", digest, filepath.Base(name))
}
