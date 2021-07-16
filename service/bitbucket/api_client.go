package bitbucket

import (
	"context"

	"github.com/reviewdog/reviewdog"
)

type ReportRequest struct {
	Owner      string
	Repository string
	Commit     string
	ReportID   string
	Type       string
	Title      string
	Reporter   string
	Result     string
	Details    string
	LogoURL    string
}

type AnnotationsRequest struct {
	Owner      string
	Repository string
	Commit     string
	ReportID   string
	Comments   []*reviewdog.Comment
}

type APIClient interface {
	CreateOrUpdateReport(ctx context.Context, req *ReportRequest) error
	CreateOrUpdateAnnotations(ctx context.Context, req *AnnotationsRequest) error
}
