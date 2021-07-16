package bitbucket

import (
	"context"
	"testing"

	"github.com/reviewdog/reviewdog/filter"
	"github.com/reviewdog/reviewdog/proto/rdf"

	"github.com/reviewdog/reviewdog"

	"github.com/stretchr/testify/suite"
)

type AnnotatorTestSuite struct {
	suite.Suite
	cli   *MockAPIClient
	owner string
	repo  string
	sha   string
}

func (s *AnnotatorTestSuite) SetupTest() {
	s.cli = &MockAPIClient{}
	s.owner = "test-owner"
	s.repo = "test-repo"
	s.sha = "test-commit"
}

// Empty runners list, no comments
func (s *AnnotatorTestSuite) TestEmptyRunnersList() {
	ctx, annotator := s.createAnnotator(nil)

	err := annotator.Flush(ctx)
	s.Require().NoError(err)
	s.cli.AssertExpectations(s.T())
}

// Predefined runners list, and no comments
func (s *AnnotatorTestSuite) TestNoComments() {
	runners := []string{"runner1", "runner2"}
	ctx, annotator := s.createAnnotator(runners)

	for _, runner := range runners {
		s.assumeReportCreated(ctx, runner, reportResultPassed)
	}

	err := annotator.Flush(ctx)

	s.Require().NoError(err)
	s.cli.AssertExpectations(s.T())
}

// Predefined runners list, and one comment
func (s *AnnotatorTestSuite) TestOneComment() {
	runners := []string{"runner1", "runner2"}
	comments := []*reviewdog.Comment{
		s.buildComment(runners[1], "main.go", "test", 1),
	}

	ctx, annotator := s.createAnnotator(runners)
	s.setupExpectedAPICalls(ctx, runners, comments)

	for _, comment := range comments {
		err := annotator.Post(ctx, comment)
		s.Require().NoError(err)
	}

	err := annotator.Flush(ctx)

	s.Require().NoError(err)
	s.cli.AssertExpectations(s.T())
}

// Predefined runners list, and duplicated comment
func (s *AnnotatorTestSuite) TestDuplicateComments() {
	runners := []string{"runner1", "runner2"}
	comments := []*reviewdog.Comment{
		s.buildComment(runners[1], "main.go", "test", 1),
		s.buildComment(runners[1], "main.go", "test", 1),
	}

	ctx, annotator := s.createAnnotator(runners)
	s.setupExpectedAPICalls(ctx, runners, comments)

	for _, comment := range comments {
		err := annotator.Post(ctx, comment)
		s.Require().NoError(err)
	}

	err := annotator.Flush(ctx)

	s.Require().NoError(err)
	s.cli.AssertExpectations(s.T())
}

// Predefined runners list, and duplicated comment
func (s *AnnotatorTestSuite) TestManyComments() {
	runners := []string{"runner1", "runner2"}

	comments := make([]*reviewdog.Comment, 333)
	for idx := 0; idx < 333; idx++ {
		comments[idx] = s.buildComment(runners[1], "main.go", "test", int32(idx))
	}

	ctx, annotator := s.createAnnotator(runners)
	s.setupExpectedAPICalls(ctx, runners, comments)

	for _, comment := range comments {
		err := annotator.Post(ctx, comment)
		s.Require().NoError(err)
	}

	err := annotator.Flush(ctx)

	s.Require().NoError(err)
	s.cli.AssertExpectations(s.T())
}

func (s *AnnotatorTestSuite) createAnnotator(runners []string) (context.Context, *ReportAnnotator) {
	ctx := context.Background()

	for _, runner := range runners {
		s.assumeReportCreated(ctx, runner, reportResultPending)
	}

	annotator := NewReportAnnotator(s.cli, s.owner, s.repo, s.sha, runners)

	return ctx, annotator
}

func (s *AnnotatorTestSuite) setupExpectedAPICalls(
	ctx context.Context,
	runners []string,
	comments []*reviewdog.Comment,
) {
	commentsMap := s.splitComments(comments)

	for _, runner := range runners {
		expResult := reportResultPassed
		if len(commentsMap[runner]) > 0 {
			expResult = reportResultFailed
		}
		s.assumeReportCreated(ctx, runner, expResult)

		for start, annCount := 0, len(commentsMap[runner]); start < annCount; start += annotationsBatchSize {
			end := start + annotationsBatchSize

			if end > annCount {
				end = annCount
			}

			s.assumeAnnotationsCreated(ctx, runner, commentsMap[runner][start:end])
		}
	}
}

func (s *AnnotatorTestSuite) assumeReportCreated(ctx context.Context, runner string, status string) {
	s.cli.On("CreateOrUpdateReport", ctx, s.buildReportReq(runner, status)).Return(nil).Once()
}

func (s *AnnotatorTestSuite) assumeAnnotationsCreated(
	ctx context.Context,
	runner string,
	comments []*reviewdog.Comment,
) {
	s.cli.On("CreateOrUpdateAnnotations", ctx, s.buildAnnotationsRequest(runner, comments)).Return(nil).Once()
}

func (s *AnnotatorTestSuite) buildReportReq(runner string, result string) *ReportRequest {
	report := &ReportRequest{
		ReportID:   reportID(runner, reporter),
		Owner:      s.owner,
		Repository: s.repo,
		Commit:     s.sha,
		Type:       reportTypeBug,
		Title:      reportTitle(runner, reporter),
		Reporter:   reporter,
		Result:     result,
		LogoURL:    logoURL,
	}

	switch result {
	case reportResultPassed:
		report.Details = "Great news! Reviewdog couldn't spot any issues!"
	case reportResultPending:
		report.Details = "Please wait for Reviewdog to finish checking your code for issues."
	default:
		report.Details = "Woof-Woof! This report generated for you by reviewdog."
	}

	return report
}

func (s *AnnotatorTestSuite) buildAnnotationsRequest(runner string, comments []*reviewdog.Comment) *AnnotationsRequest {
	return &AnnotationsRequest{
		Owner:      s.owner,
		Repository: s.repo,
		Commit:     s.sha,
		ReportID:   reportID(runner, reporter),
		Comments:   comments,
	}
}

func (s *AnnotatorTestSuite) buildComment(toolName, file, message string, line int32) *reviewdog.Comment {
	return &reviewdog.Comment{
		ToolName: toolName,
		Result: &filter.FilteredDiagnostic{
			Diagnostic: &rdf.Diagnostic{
				Location: &rdf.Location{
					Path: file,
					Range: &rdf.Range{Start: &rdf.Position{
						Line: line,
					}},
				},
				Message: message,
			},
		},
	}
}

func (s *AnnotatorTestSuite) splitComments(comments []*reviewdog.Comment) map[string][]*reviewdog.Comment {
	commentsMap := make(map[string][]*reviewdog.Comment)
	duplicates := make(map[string]struct{})

	for _, comment := range comments {
		externalID := externalIDFromDiagnostic(comment.Result.Diagnostic)
		if _, exist := duplicates[externalID]; !exist {
			commentsMap[comment.ToolName] = append(commentsMap[comment.ToolName], comment)
			duplicates[externalID] = struct{}{}
		}
	}

	return commentsMap
}

func TestAnnotatorTestSuite(t *testing.T) {
	suite.Run(t, &AnnotatorTestSuite{})
}
