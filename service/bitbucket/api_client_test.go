package bitbucket

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) CreateOrUpdateReport(ctx context.Context, report *ReportRequest) error {
	return m.Called(ctx, report).Error(0)
}

func (m *MockAPIClient) CreateOrUpdateAnnotations(ctx context.Context, annotations *AnnotationsRequest) error {
	return m.Called(ctx, annotations).Error(0)
}
