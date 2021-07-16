package bitbucket

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/reviewdog/reviewdog"
)

var _ reviewdog.CommentService = &ReportAnnotator{}

const (
	// avatar from https://github.com/apps/reviewdog
	logoURL  = "https://avatars1.githubusercontent.com/in/12131"
	reporter = "reviewdog"
	// max amount of annotations in one batch call
	annotationsBatchSize = 100
)

// ReportAnnotator is a comment service for Bitbucket Code Insights reports.
//
// Cloud API:
//  https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/commit/%7Bcommit%7D/reports/%7BreportId%7D/annotations#post
//  POST /2.0/repositories/{username}/{repo_slug}/commit/{commit}/reports/{reportId}/annotations
//
// Server API:
//  https://docs.atlassian.com/bitbucket-server/rest/5.15.0/bitbucket-code-insights-rest.html#idm288218233536
//  /rest/insights/1.0/projects/{projectKey}/repos/{repositorySlug}/commits/{commitId}/reports/{key}/annotations
type ReportAnnotator struct {
	cli         APIClient
	sha         string
	owner, repo string

	muAnnotations sync.Mutex
	// store annotations in map per tool name
	// so we can create report per tool
	comments map[string][]*reviewdog.Comment

	// wd is working directory relative to root of repository.
	wd         string
	duplicates map[string]struct{}
}

// NewReportAnnotator creates new Bitbucket ReportRequest Annotator
func NewReportAnnotator(cli APIClient, owner, repo, sha string, runners []string) *ReportAnnotator {
	r := &ReportAnnotator{
		cli:        cli,
		sha:        sha,
		owner:      owner,
		repo:       repo,
		comments:   make(map[string][]*reviewdog.Comment, len(runners)),
		duplicates: map[string]struct{}{},
	}

	// pre populate map of annotations, so we still create passed (green) report
	// if no issues found from the specific tool
	for _, runner := range runners {
		if len(runner) == 0 {
			continue
		}
		r.comments[runner] = []*reviewdog.Comment{}
		// create Pending report for each tool
		_ = r.createOrUpdateReport(
			context.Background(),
			reportID(runner, reporter),
			reportTitle(runner, reporter),
			reportResultPending,
		)
	}

	return r
}

// Post accepts a comment and holds it. Flush method actually posts comments to
// Bitbucket in batch.
func (r *ReportAnnotator) Post(_ context.Context, c *reviewdog.Comment) error {
	c.Result.Diagnostic.GetLocation().Path = filepath.ToSlash(
		filepath.Join(r.wd, c.Result.Diagnostic.GetLocation().GetPath()))

	r.muAnnotations.Lock()
	defer r.muAnnotations.Unlock()

	// deduplicate event, because some reporters might report
	// it twice, and bitbucket api will complain on duplicated
	// external id of annotation
	commentID := externalIDFromDiagnostic(c.Result.Diagnostic)
	if _, exist := r.duplicates[commentID]; !exist {
		r.comments[c.ToolName] = append(r.comments[c.ToolName], c)
		r.duplicates[commentID] = struct{}{}
	}

	return nil
}

// Flush posts comments which has not been posted yet.
func (r *ReportAnnotator) Flush(ctx context.Context) error {
	r.muAnnotations.Lock()
	defer r.muAnnotations.Unlock()

	// create/update/annotate report per tool
	for tool, comments := range r.comments {
		reportID := reportID(tool, reporter)
		title := reportTitle(tool, reporter)
		if len(comments) == 0 {
			// if no annotation, create Passed report
			if err := r.createOrUpdateReport(ctx, reportID, title, reportResultPassed); err != nil {
				return err
			}
			// and move one
			continue
		}

		// create report or update report first, with the failed status
		if err := r.createOrUpdateReport(ctx, reportID, title, reportResultFailed); err != nil {
			return err
		}

		// send comments in batches, because of the api max payload size limit
		for start, annCount := 0, len(comments); start < annCount; start += annotationsBatchSize {
			end := start + annotationsBatchSize

			if end > annCount {
				end = annCount
			}

			req := &AnnotationsRequest{
				Owner:      r.owner,
				Repository: r.repo,
				Commit:     r.sha,
				ReportID:   reportID,
				Comments:   comments[start:end],
			}

			err := r.cli.CreateOrUpdateAnnotations(ctx, req)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ReportAnnotator) createOrUpdateReport(ctx context.Context, id, title, reportStatus string) error {
	req := &ReportRequest{
		ReportID:   id,
		Owner:      r.owner,
		Repository: r.repo,
		Commit:     r.sha,
		Type:       reportTypeBug,
		Title:      title,
		Reporter:   reporter,
		Result:     reportStatus,
		LogoURL:    logoURL,
	}

	switch reportStatus {
	case reportResultPassed:
		req.Details = "Great news! Reviewdog couldn't spot any issues!"
	case reportResultPending:
		req.Details = "Please wait for Reviewdog to finish checking your code for issues."
	default:
		req.Details = "Woof-Woof! This report generated for you by reviewdog."
	}

	return r.cli.CreateOrUpdateReport(ctx, req)
}
