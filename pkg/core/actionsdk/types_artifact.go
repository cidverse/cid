package actionsdk

import (
	"github.com/cidverse/cid/pkg/util"
)

type Artifact struct {
	BuildID       string `json:"build_id"`
	JobID         string `json:"job_id"`
	StepSlug      string `json:"step_slug"`
	ArtifactID    string `json:"id"`
	Module        string `json:"module"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Format        string `json:"format"`
	FormatVersion string `json:"format_version"`
	SHA256        string `json:"sha256"`
}

type ArtifactListRequest struct {
	Query string `json:"query"`
}

type ArtifactUploadRequest struct {
	File          string `json:"file"`
	Content       string `json:"content"`
	ContentBytes  []byte `json:"content_bytes"`
	Module        string `json:"module"`
	Type          string `json:"type"`
	Format        string `json:"format"`
	FormatVersion string `json:"format_version"`
	ExtractFile   bool   `json:"extract_file"`
}

func (r ArtifactUploadRequest) ModuleSlug() string {
	return util.GetStringOrDefault(r.Module, "root")
}

type ArtifactDownloadRequest struct {
	ID         string `json:"id"`
	TargetFile string `json:"target_file"`
}

type ArtifactDownloadResult struct {
	Path string
	Hash string
	Size int64
}

type ArtifactDownloadByteArrayRequest struct {
	ID string `json:"id"`
}

type ArtifactDownloadByteArrayResult struct {
	Bytes []byte
	Hash  string
	Size  int64
}
