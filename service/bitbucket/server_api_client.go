package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	insights "github.com/reva2/bitbucket-insights-api"
)

type ServerAPIClient struct {
	cli    *insights.APIClient
	helper *ServerAPIHelper
}

// NewServerAPIClient creates client for Bitbucket Server
func NewServerAPIClient() APIClient {
	httpClient := &http.Client{
		Timeout: httpTimeout,
	}

	config := insights.NewConfiguration()
	config.HTTPClient = httpClient

	return &ServerAPIClient{
		cli:    insights.NewAPIClient(config),
		helper: &ServerAPIHelper{},
	}
}

func (c *ServerAPIClient) CreateOrUpdateReport(ctx context.Context, req *ReportRequest) error {
	// Bitbucket Server API doesn't support pending status
	if req.Result == reportResultPending {
		return nil
	}

	// We need to delete all annotations created previously
	err := c.deleteReport(ctx, req)
	if err != nil {
		return err
	}

	_, resp, err := c.cli.InsightsApi.
		UpdateReport(ctx, req.Owner, req.Repository, req.Commit, req.ReportID).
		Report(c.helper.BuildReport(req)).
		Execute()

	if err := c.checkAPIError(err, resp, http.StatusOK); err != nil {
		return fmt.Errorf("insights.UpdateReport: %s", err)
	}

	return nil
}

func (c *ServerAPIClient) CreateOrUpdateAnnotations(ctx context.Context, req *AnnotationsRequest) error {
	resp, err := c.cli.InsightsApi.
		CreateAnnotations(ctx, req.Owner, req.Repository, req.Commit, req.ReportID).
		AnnotationsList(c.helper.BuildAnnotations(req.Comments)).
		Execute()

	if err := c.checkAPIError(err, resp, http.StatusNoContent); err != nil {
		return fmt.Errorf("insights.CreateAnnotations: %s", err)
	}

	return nil
}

func (c *ServerAPIClient) deleteReport(ctx context.Context, report *ReportRequest) error {
	resp, err := c.cli.InsightsApi.
		DeleteReport(ctx, report.Owner, report.Repository, report.Commit, report.ReportID).
		Execute()

	if err := c.checkAPIError(err, resp, http.StatusNoContent); err != nil {
		return fmt.Errorf("insights.DeleteReport: %s", err)
	}

	return nil
}

func (c *ServerAPIClient) checkAPIError(err error, resp *http.Response, expectedCode int) error {
	if err != nil {
		e, ok := err.(insights.GenericOpenAPIError)
		if ok {
			return fmt.Errorf(`bitbucket API error:
	Response error: %s
	Response body: %s`,
				e.Error(), string(e.Body()))
		}
	}

	if resp != nil && resp.StatusCode != expectedCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("received unexpected %d code from Bitbucket API", resp.StatusCode)
		if len(body) > 0 {
			msg += " with message:\n" + string(body)
		}
		return errors.New(msg)
	}

	return nil
}
