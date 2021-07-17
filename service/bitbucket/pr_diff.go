package bitbucket

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	bbv1api "github.com/gfleury/go-bitbucket-v1"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/service/serviceutil"
)

var _ reviewdog.DiffService = &PullRequestDiff{}

// PullRequestDiff is a diff service for BitBucket PullRequest.
type PullRequestDiff struct {
	cli              *bbv1api.APIClient
	pr               int
	sha, owner, repo string

	// wd is working directory relative to root of repository.
	wd string
}

// NewPullRequestDiff returns a new PullRequestDiff service.
// PullRequestDiff service needs git command in $PATH.
func NewPullRequestDiff(cli *bbv1api.APIClient, owner, repo string, pr int, sha string) (*PullRequestDiff, error) {
	workDir, err := serviceutil.GitRelWorkdir()
	if err != nil {
		return nil, fmt.Errorf("PullRequestDiff needs 'git' command: %w", err)
	}
	return &PullRequestDiff{
		cli:   cli,
		pr:    pr,
		sha:   sha,
		owner: owner,
		repo:  repo,
		wd:    workDir,
	}, nil
}

// Diff returns a diff of PullRequest. It runs `git diff` locally instead of
// diff of BitBucket Pull Request to avoid converting json representation of
// BitBucket diff to unified diff format
func (g *PullRequestDiff) Diff(ctx context.Context) ([]byte, error) {
	response, err := g.cli.DefaultApi.GetPullRequest(g.owner, g.repo, g.pr)
	if err != nil {
		return nil, err
	}
	pr, err := bbv1api.GetPullRequestResponse(response)
	if err != nil {
		return nil, err
	}
	return g.gitDiff(ctx, g.sha, pr.ToRef.LatestCommit)
}

func (g *PullRequestDiff) gitDiff(_ context.Context, baseSha, targetSha string) ([]byte, error) {
	b, err := exec.Command("git", "merge-base", targetSha, baseSha).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get merge-base commit: %w", err)
	}
	mergeBase := strings.Trim(string(b), "\n")
	bytes, err := exec.Command("git", "diff", "--find-renames", mergeBase, baseSha).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff: %w", err)
	}
	return bytes, nil
}

// Strip returns 1 as a strip of git diff.
func (g *PullRequestDiff) Strip() int {
	return 1
}
