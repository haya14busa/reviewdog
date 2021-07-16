package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	bbapi "github.com/reviewdog/go-bitbucket"
)

const (
	// PipelineProxyURL available while using Bitbucket Pipelines and
	// allows you to use the Reports-API without extra authentication.
	// For that you need to send your request through a proxy server that runs alongside with
	// every pipeline on ‘localhost:29418’, and a valid Auth-Header will automatically be added to your request.
	// https://support.atlassian.com/bitbucket-cloud/docs/code-insights/#Authentication
	// However, if using proxy HTTP API endpoint need to be used
	pipelineProxyURL = "http://localhost:29418"
)

type CloudAPIClient struct {
	cli    *bbapi.APIClient
	helper *CloudAPIHelper
}

// NewCloudAPIClient creates client for Bitbucket Cloud
func NewCloudAPIClient(isInPipeline bool) APIClient {
	httpClient := &http.Client{
		Timeout: httpTimeout,
	}

	server := bbapi.ServerConfiguration{
		URL:         "https://api.bitbucket.org/2.0",
		Description: `HTTPS API endpoint`,
	}

	if isInPipeline {
		// if we are on the Bitbucket Pipeline, use HTTP endpoint
		// and proxy
		proxyURL, _ := url.Parse(pipelineProxyURL)
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		server = bbapi.ServerConfiguration{
			URL:         "http://api.bitbucket.org/2.0",
			Description: `If if called from Bitbucket Pipelines, using HTTP API endpoint and AuthProxy`,
		}
	}

	return NewCloudAPIClientWithConfigurations(httpClient, server)
}

func NewCloudAPIClientWithConfigurations(client *http.Client, server bbapi.ServerConfiguration) APIClient {
	config := bbapi.NewConfiguration()
	if client != nil {
		config.HTTPClient = client
	} else {
		config.HTTPClient = &http.Client{
			Timeout: httpTimeout,
		}
	}
	config.Servers = bbapi.ServerConfigurations{server}

	return &CloudAPIClient{
		cli:    bbapi.NewAPIClient(config),
		helper: &CloudAPIHelper{},
	}
}

func (c *CloudAPIClient) CreateOrUpdateReport(ctx context.Context, req *ReportRequest) error {
	_, resp, err := c.cli.
		ReportsApi.CreateOrUpdateReport(ctx, req.Owner, req.Repository, req.Commit, req.ReportID).
		Body(c.helper.BuildReport(req)).
		Execute()

	if err := c.checkAPIError(err, resp, http.StatusOK); err != nil {
		return fmt.Errorf("bitbucket.CreateOrUpdateReport: %s", err)
	}

	return nil
}

func (c *CloudAPIClient) CreateOrUpdateAnnotations(ctx context.Context, req *AnnotationsRequest) error {
	_, resp, err := c.cli.ReportsApi.
		BulkCreateOrUpdateAnnotations(ctx, req.Owner, req.Repository, req.Commit, req.ReportID).
		Body(c.helper.BuildAnnotations(req.Comments)).
		Execute()

	if err := c.checkAPIError(err, resp, http.StatusOK); err != nil {
		return fmt.Errorf("bitbucket.BulkCreateOrUpdateAnnotations: %s", err)
	}

	return nil
}

func (c *CloudAPIClient) checkAPIError(err error, resp *http.Response, expectedCode int) error {
	if err != nil {
		e, ok := err.(bbapi.GenericOpenAPIError)
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
