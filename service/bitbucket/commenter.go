package bitbucket

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	bbv1api "github.com/gfleury/go-bitbucket-v1"
	"golang.org/x/sync/errgroup"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/service/commentutil"
	"github.com/reviewdog/reviewdog/service/serviceutil"
)

var _ reviewdog.CommentService = &PullRequestCommenter{}

// PullRequestCommenter is a comment service for Bitbucket pull request discussion.
//
// API:
//  https://docs.atlassian.com/bitbucket-server/rest/5.16.0/bitbucket-rest.html#idm8286336848
//  POST /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}/pull-requests/{pullRequestId}/comments
type PullRequestCommenter struct {
	cli              *bbv1api.APIClient
	pr               int64
	sha, owner, repo string

	muComments   sync.Mutex
	postComments []*reviewdog.Comment

	// wd is working directory relative to root of repository.
	wd string
}

// NewPullRequestCommenter returns a new PullRequestCommenter service.
// PullRequestCommenter service needs git command in $PATH.
func NewPullRequestCommenter(cli *bbv1api.APIClient, owner, repo string, pr int, sha string) (*PullRequestCommenter,
	error) {
	workDir, err := serviceutil.GitRelWorkdir()
	if err != nil {
		return nil, fmt.Errorf("PullRequestCommenter needs 'git' command: %w", err)
	}
	return &PullRequestCommenter{
		cli:   cli,
		pr:    int64(pr),
		sha:   sha,
		owner: owner,
		repo:  repo,
		wd:    workDir,
	}, nil
}

// Post accepts a comment and holds it. Flush method actually posts comments to
// BitBucket in parallel.
func (g *PullRequestCommenter) Post(_ context.Context, c *reviewdog.Comment) error {
	c.Result.Diagnostic.GetLocation().Path = filepath.ToSlash(
		filepath.Join(g.wd, c.Result.Diagnostic.GetLocation().GetPath()))
	g.muComments.Lock()
	defer g.muComments.Unlock()
	g.postComments = append(g.postComments, c)
	return nil
}

// Flush posts comments which has not been posted yet.
func (g *PullRequestCommenter) Flush(ctx context.Context) error {
	g.muComments.Lock()
	defer g.muComments.Unlock()
	postedcs, err := g.createPostedComments()
	if err != nil {
		return fmt.Errorf("failed to create posted comments: %w", err)
	}
	return g.postCommentsForEach(ctx, postedcs)
}

func (g *PullRequestCommenter) createPostedComments() (commentutil.PostedComments, error) {
	postedcs := make(commentutil.PostedComments)
	activities, err := listAllPullRequestActivities(g.cli, g.owner, g.repo, g.pr, map[string]interface{}{"limit": 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list all pull request activities: %w", err)
	}
	for _, a := range activities {
		if a.Action != bbv1api.ActionCommented || a.CommentAnchor.Line == 0 || a.Comment.Text == "" {
			continue
		}
		postedcs.AddPostedComment(a.CommentAnchor.Path, a.CommentAnchor.Line, a.Comment.Text)
	}
	return postedcs, nil
}

func (g *PullRequestCommenter) postCommentsForEach(_ context.Context, postedcs commentutil.PostedComments) error {
	var eg errgroup.Group
	for _, c := range g.postComments {
		c := c
		loc := c.Result.Diagnostic.GetLocation()
		lnum := int(loc.GetRange().GetStart().GetLine())
		body := commentutil.BitBucketMarkdownComment(c)
		if !c.Result.InDiffFile || lnum == 0 || postedcs.IsPosted(c, lnum, body) {
			continue
		}
		eg.Go(func() error {
			anchor := &bbv1api.Anchor{
				DiffType: bbv1api.DiffTypeEffective,
				Line:     lnum,
				LineType: bbv1api.LineTypeAdded,
				FileType: bbv1api.FileTypeTo,
				Path:     loc.GetPath(),
				SrcPath:  c.Result.OldPath,
			}
			comment := bbv1api.Comment{
				Text:   body,
				Anchor: anchor,
			}
			_, err := g.cli.DefaultApi.CreatePullRequestComment(g.owner, g.repo, int(g.pr), comment,
				[]string{"application/json"})
			if err != nil {
				return fmt.Errorf("failed to create pull request comment: %w", err)
			}
			return nil
		})
	}
	return eg.Wait()
}

func listAllPullRequestActivities(cli *bbv1api.APIClient, owner, repo string, pr int64, opts map[string]interface{}) (
	[]bbv1api.Activity, error) {
	resp, err := cli.DefaultApi.GetActivities(owner, repo, pr, opts)
	if err != nil {
		return nil, err
	}
	activities, err := bbv1api.GetActivitiesResponse(resp)
	if err != nil {
		return nil, err
	}
	if activities.IsLastPage {
		return activities.Values, nil
	}
	newOpts := map[string]interface{}{
		"start": activities.NextPageStart,
		"limit": opts["limit"],
	}
	restActivities, err := listAllPullRequestActivities(cli, owner, repo, pr, newOpts)
	if err != nil {
		return nil, err
	}
	return append(activities.Values, restActivities...), nil
}
