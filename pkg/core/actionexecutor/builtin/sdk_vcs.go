package builtin

import (
	"fmt"
	"sort"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/go-vcs"
	"github.com/cidverse/go-vcs/vcsapi"
	"github.com/hashicorp/go-version"
)

func (sdk ActionSDK) VCSCommitsV1(request actionsdk.VCSCommitsRequest) ([]*actionsdk.VCSCommit, error) {
	fromRef, err := vcsapi.NewVCSRefFromString(request.FromHash)
	if err != nil {
		return nil, fmt.Errorf("parameter has a invalid value: from: %w", err)
	}
	toRef, err := vcsapi.NewVCSRefFromString(request.ToHash)
	if err != nil {
		return nil, fmt.Errorf("parameter has a invalid value: to: %w", err)
	}

	client, err := vcs.GetVCSClient(sdk.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open vcs repository: %w", err)
	}

	commits, err := client.FindCommitsBetween(fromRef, toRef, request.IncludeChanges, request.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query commits: %w", err)
	}

	return convertToVCSCommits(commits), nil
}

func (sdk ActionSDK) VCSCommitByHashV1(request actionsdk.VCSCommitByHashRequest) (*actionsdk.VCSCommit, error) {
	client, clientErr := vcs.GetVCSClient(sdk.ProjectDir)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to open vcs repository: %w", clientErr)
	}

	commit, err := client.FindCommitByHash(request.Hash, request.IncludeChanges)
	if err != nil {
		return nil, fmt.Errorf("failed to query commit: %w", err)
	}

	c := convertToVCSCommit(commit)
	return &c, nil
}

func (sdk ActionSDK) VCSTagsV1() ([]actionsdk.VCSTag, error) {
	client, clientErr := vcs.GetVCSClient(sdk.ProjectDir)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to open vcs repository: %w", clientErr)
	}

	tags := client.GetTags()

	var result []actionsdk.VCSTag
	for _, tag := range tags {
		result = append(result, actionsdk.VCSTag{
			RefType: tag.Type,
			Value:   tag.Value,
			Hash:    tag.Hash,
		})
	}
	return result, nil
}

func (sdk ActionSDK) VCSReleasesV1(request actionsdk.VCSReleasesRequest) ([]actionsdk.VCSRelease, error) {
	releaseType := request.Type

	client, clientErr := vcs.GetVCSClient(sdk.ProjectDir)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to open vcs repository: %w", clientErr)
	}

	var versions []*version.Version
	var versionToTag = make(map[string]vcsapi.VCSRef)
	for _, tag := range client.GetTags() {
		v, vErr := version.NewVersion(tag.Value)
		if vErr == nil {
			versions = append(versions, v)
			versionToTag[v.String()] = tag
		}
	}
	sort.Sort(util.ByVersion(versions))

	var releases []actionsdk.VCSRelease
	for _, v := range versions {
		release := actionsdk.VCSRelease{
			Version: v.String(),
			Ref:     versionToTag[v.String()],
		}

		if len(releaseType) > 0 {
			if releaseType == "stable" && len(v.Prerelease()) > 0 {
				continue
			} else if releaseType == "unstable" && v.Prerelease() == "" {
				continue
			} else if releaseType != "stable" && releaseType != "unstable" {
				return nil, fmt.Errorf("release type must be empty, stable or unstable")
			}
		}

		releases = append(releases, release)
	}

	return releases, nil
}

func (sdk ActionSDK) VCSDiffV1(request actionsdk.VCSDiffRequest) ([]actionsdk.VCSDiff, error) {
	fromRef, err := vcsapi.NewVCSRefFromString(request.FromHash)
	if err != nil {
		return nil, fmt.Errorf("parameter has a invalid value: from: %w", err)
	}

	toRef, err := vcsapi.NewVCSRefFromString(request.ToHash)
	if err != nil {
		return nil, fmt.Errorf("parameter has a invalid value: to: %w", err)
	}

	client, err := vcs.GetVCSClient(sdk.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open vcs repository: %w", err)
	}

	diff, err := client.Diff(fromRef, toRef)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff: %w", err)
	}

	return convertToVCSDiff(diff), nil
}

func convertToVCSCommits(commits []vcsapi.Commit) []*actionsdk.VCSCommit {
	result := make([]*actionsdk.VCSCommit, len(commits))

	for i, c := range commits {
		r := convertToVCSCommit(c)
		result[i] = &r
	}

	return result
}

func convertToVCSCommit(commit vcsapi.Commit) actionsdk.VCSCommit {
	return actionsdk.VCSCommit{
		HashShort:   commit.ShortHash,
		Hash:        commit.Hash,
		Message:     commit.Message,
		Description: commit.Description,
		Author:      convertToVCSAuthor(commit.Author),
		Committer:   convertToVCSAuthor(commit.Committer),
		Tags:        convertToVCSTags(commit.Tags),
		AuthoredAt:  commit.AuthoredAt,
		CommittedAt: commit.CommittedAt,
		Changes:     convertToVCSChange(commit.Changes),
		Context:     commit.Context,
	}
}

func convertToVCSAuthor(author vcsapi.CommitAuthor) actionsdk.VCSAuthor {
	return actionsdk.VCSAuthor{
		Name:  author.Name,
		Email: author.Email,
	}
}

func convertToVCSTags(tag []vcsapi.VCSRef) []*actionsdk.VCSTag {
	tags := make([]*actionsdk.VCSTag, len(tag))

	for i, t := range tag {
		tags[i] = &actionsdk.VCSTag{
			RefType: t.Type,
			Value:   t.Value,
			Hash:    t.Hash,
		}
	}

	return tags
}

func convertToVCSChange(changes []vcsapi.CommitChange) []*actionsdk.VCSChange {
	result := make([]*actionsdk.VCSChange, len(changes))

	for i, c := range changes {
		result[i] = &actionsdk.VCSChange{
			ChangeType: c.Type,
			FileFrom: actionsdk.VCSFile{
				Name: c.FileFrom.Name,
				Size: int(c.FileFrom.Size),
				Hash: c.FileFrom.Hash,
			},
			FileTo: actionsdk.VCSFile{
				Name: c.FileTo.Name,
				Size: int(c.FileTo.Size),
				Hash: c.FileTo.Hash,
			},
			Patch: c.Patch,
		}
	}

	return result
}

func convertToVCSDiff(diff []vcsapi.VCSDiff) []actionsdk.VCSDiff {
	result := make([]actionsdk.VCSDiff, len(diff))

	for i, d := range diff {
		result[i] = actionsdk.VCSDiff{
			FileFrom: actionsdk.VCSFile{
				Name: d.FileFrom.Name,
				Size: int(d.FileFrom.Size),
				Hash: d.FileFrom.Hash,
			},
			FileTo: actionsdk.VCSFile{
				Name: d.FileTo.Name,
				Size: int(d.FileTo.Size),
				Hash: d.FileTo.Hash,
			},
			Lines: convertToVCSDiffLines(d.Lines),
		}
	}

	return result
}

func convertToVCSDiffLines(lines []vcsapi.VCSDiffLine) []actionsdk.VCSDiffLine {
	result := make([]actionsdk.VCSDiffLine, len(lines))

	for i, l := range lines {
		result[i] = actionsdk.VCSDiffLine{
			Operation: l.Operation,
			Content:   l.Content,
		}
	}

	return result
}
