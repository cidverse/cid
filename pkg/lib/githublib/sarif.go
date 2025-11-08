package githublib

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/cidverse/cidverseutils/compress"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	"github.com/google/go-github/v78/github"
	"golang.org/x/oauth2"
)

func GitHubCodeSecuritySarifUpload(githubToken string, sarifFile string, nci nci.Spec) error {
	parts := strings.Split(nci.Project.Path, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository path: %s", nci.Project.Path)
	}
	organization := parts[0]
	repository := parts[1]

	// GitHub Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	// read sarif file
	reportName := path.Base(sarifFile)
	sarifBytes, err := os.ReadFile(sarifFile)
	if err != nil {
		return fmt.Errorf("failed to read sarif report: %w", err)
	}

	// encoding
	sarifEncoded, reportErr := compress.GZIPBase64EncodeBytes(sarifBytes)
	if reportErr != nil {
		return fmt.Errorf("failed to encode sarif report (gzip/base64): %w", err)
	}

	// git reference (sarif upload with pull request ref will result in pull request comments)
	ref := nci.Commit.RefVCS
	if nci.Pipeline.Trigger == "merge_request" && nci.MergeRequest.Id != "" {
		ref = fmt.Sprintf("refs/pull/%s/merge", nci.MergeRequest.Id)
	}

	// upload
	slog.With("report", reportName).With("ref", ref).With("commit_hash", nci.Commit.Hash).Info("uploading sarif report to github code scanning api")
	sarifAnalysis := &github.SarifAnalysis{CommitSHA: github.Ptr(nci.Commit.Hash), Ref: github.Ptr(ref), Sarif: github.Ptr(sarifEncoded), CheckoutURI: github.Ptr(nci.Project.Dir)}
	sarifId, _, reportErr := client.CodeScanning.UploadSarif(context.Background(), organization, repository, sarifAnalysis)

	if reportErr != nil {
		// "job scheduled on GitHub side" is not an error, job just isn't completed yet
		if strings.Contains(reportErr.Error(), "job scheduled on GitHub side") {
			slog.With("report", reportName).With("state", "github_job_pending").Info("sarif upload successful")
		} else {
			return fmt.Errorf("failed to upload sarif to github code-scanning api: %s", reportErr.Error())
		}
	} else if sarifId != nil {
		slog.With("report", reportName).With("state", "ok").With("id", *sarifId.ID).With("url", *sarifId.URL).Info("sarif upload successful")
	}

	return nil
}
