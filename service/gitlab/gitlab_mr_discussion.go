package gitlab

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/sync/errgroup"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/service/commentutil"
	"github.com/reviewdog/reviewdog/service/serviceutil"
)

// MergeRequestDiscussionCommenter is a comment and diff service for GitLab MergeRequest.
//
// API:
//  https://docs.gitlab.com/ee/api/discussions.html#create-new-merge-request-discussion
//  POST /projects/:id/merge_requests/:merge_request_iid/discussions
type MergeRequestDiscussionCommenter struct {
	cli      *gitlab.Client
	pr       int
	sha      string
	projects string

	muComments   sync.Mutex
	postComments []*reviewdog.Comment

	// wd is working directory relative to root of repository.
	wd string
}

// NewGitLabMergeRequestDiscussionCommenter returns a new MergeRequestDiscussionCommenter service.
// MergeRequestDiscussionCommenter service needs git command in $PATH.
func NewGitLabMergeRequestDiscussionCommenter(cli *gitlab.Client, owner, repo string, pr int, sha string) (*MergeRequestDiscussionCommenter, error) {
	workDir, err := serviceutil.GitRelWorkdir()
	if err != nil {
		return nil, fmt.Errorf("MergeRequestDiscussionCommenter needs 'git' command: %w", err)
	}
	return &MergeRequestDiscussionCommenter{
		cli:      cli,
		pr:       pr,
		sha:      sha,
		projects: owner + "/" + repo,
		wd:       workDir,
	}, nil
}

// Post accepts a comment and holds it. Flush method actually posts comments to
// GitLab in parallel.
func (g *MergeRequestDiscussionCommenter) Post(_ context.Context, c *reviewdog.Comment) error {
	c.Result.Diagnostic.GetLocation().Path = filepath.ToSlash(
		filepath.Join(g.wd, c.Result.Diagnostic.GetLocation().GetPath()))
	g.muComments.Lock()
	defer g.muComments.Unlock()
	g.postComments = append(g.postComments, c)
	return nil
}

// Flush posts comments which has not been posted yet.
func (g *MergeRequestDiscussionCommenter) Flush(ctx context.Context) error {
	g.muComments.Lock()
	defer g.muComments.Unlock()
	postedcs, err := g.createPostedComments()
	if err != nil {
		return fmt.Errorf("failed to create posted comments: %w", err)
	}
	return g.postCommentsForEach(ctx, postedcs)
}

func (g *MergeRequestDiscussionCommenter) createPostedComments() (commentutil.PostedComments, error) {
	postedcs := make(commentutil.PostedComments)
	discussions, err := listAllMergeRequestDiscussion(g.cli, g.projects, g.pr, &gitlab.ListMergeRequestDiscussionsOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to list all merge request discussions: %w", err)
	}
	for _, d := range discussions {
		for _, note := range d.Notes {
			pos := note.Position
			if pos == nil || pos.NewPath == "" || pos.NewLine == 0 || note.Body == "" {
				continue
			}
			postedcs.AddPostedComment(pos.NewPath, pos.NewLine, note.Body)
		}
	}
	return postedcs, nil
}

func (g *MergeRequestDiscussionCommenter) postCommentsForEach(ctx context.Context, postedcs commentutil.PostedComments) error {
	mr, _, err := g.cli.MergeRequests.GetMergeRequest(g.projects, g.pr, nil, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to get merge request: %w", err)
	}
	targetBranch, _, err := g.cli.Branches.GetBranch(mr.TargetProjectID, mr.TargetBranch, nil)
	if err != nil {
		return err
	}

	var eg errgroup.Group
	for _, c := range g.postComments {
		c := c
		loc := c.Result.Diagnostic.GetLocation()
		lnum := int(loc.GetRange().GetStart().GetLine())
		body := commentutil.MarkdownComment(c)

		if suggestion := commentutil.MarkdownSuggestions(c); suggestion != "" {
			body = body + "\n\n" + suggestion
		}

		if !c.Result.InDiffFile || lnum == 0 || postedcs.IsPosted(c, lnum, body) {
			continue
		}
		eg.Go(func() error {
			pos := &gitlab.NotePosition{
				StartSHA:     targetBranch.Commit.ID,
				HeadSHA:      g.sha,
				BaseSHA:      targetBranch.Commit.ID,
				PositionType: "text",
				NewPath:      loc.GetPath(),
				NewLine:      lnum,
			}
			if c.Result.OldPath != "" && c.Result.OldLine != 0 {
				pos.OldPath = c.Result.OldPath
				pos.OldLine = c.Result.OldLine
			}
			discussion := &gitlab.CreateMergeRequestDiscussionOptions{
				Body:     gitlab.String(body),
				Position: pos,
			}
			_, _, err := g.cli.Discussions.CreateMergeRequestDiscussion(g.projects, g.pr, discussion)
			if err != nil {
				return fmt.Errorf("failed to create merge request discussion: %w", err)
			}
			return nil
		})
	}
	return eg.Wait()
}

func listAllMergeRequestDiscussion(cli *gitlab.Client, projectID string, mergeRequest int, opts *gitlab.ListMergeRequestDiscussionsOptions) ([]*gitlab.Discussion, error) {
	discussions, resp, err := cli.Discussions.ListMergeRequestDiscussions(projectID, mergeRequest, opts)
	if err != nil {
		return nil, err
	}
	if resp.NextPage == 0 {
		return discussions, nil
	}
	newOpts := &gitlab.ListMergeRequestDiscussionsOptions{
		Page:    resp.NextPage,
		PerPage: opts.PerPage,
	}
	restDiscussions, err := listAllMergeRequestDiscussion(cli, projectID, mergeRequest, newOpts)
	if err != nil {
		return nil, err
	}
	return append(discussions, restDiscussions...), nil
}
