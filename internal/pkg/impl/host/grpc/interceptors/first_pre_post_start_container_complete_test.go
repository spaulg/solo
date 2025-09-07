package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpc_mock "github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/container"
)

func TestFirstPreStartCompleteTestSuite(t *testing.T) {
	suite.Run(t, new(FirstPreStartCompleteTestSuite))
}

type FirstPreStartCompleteTestSuite struct {
	suite.Suite

	mockOrchestrator *container.MockOrchestrator
}

func (t *FirstPreStartCompleteTestSuite) SetupTest() {
	t.mockOrchestrator = &container.MockOrchestrator{}
}

func (t *FirstPreStartCompleteTestSuite) TestSuccessfulFirstPreStartCompleteUnaryInterceptor() {
	expectedFirstPreStartComplete := "true"
	md := metadata.MD{}
	md.Set(FirstPreStartContainerCompleteMetadataKey, expectedFirstPreStartComplete)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewFirstPreStartCompleteInterceptor(t.mockOrchestrator)
	result, err := interceptor.FirstPreStartCompleteUnaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
		firstPreStartComplete, ok := ctx.Value(FirstPreStartComplete(FirstPreStartContainerCompleteContextValueName)).(string)

		t.True(ok)
		t.Equal(expectedFirstPreStartComplete, firstPreStartComplete)

		return expectedResult, nil
	})

	t.Equal(expectedResult, result)
	t.NoError(err)
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *FirstPreStartCompleteTestSuite) TestFirstPreStartCompleteUnaryInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewFirstPreStartCompleteInterceptor(t.mockOrchestrator)
	_, err := interceptor.FirstPreStartCompleteUnaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *FirstPreStartCompleteTestSuite) TestSuccessfulFirstPreStartCompleteStreamInterceptor() {
	expectedFirstPreStartComplete := "true"
	md := metadata.MD{}
	md.Set(FirstPreStartContainerCompleteMetadataKey, expectedFirstPreStartComplete)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	interceptor := NewFirstPreStartCompleteInterceptor(t.mockOrchestrator)
	err := interceptor.FirstPreStartCompleteStreamInterceptor(srv, ss, info, func(srv any, stream grpc.ServerStream) error {
		firstPreStartComplete, ok := stream.Context().Value(FirstPreStartComplete(FirstPreStartContainerCompleteContextValueName)).(string)

		t.True(ok)
		t.Equal(expectedFirstPreStartComplete, firstPreStartComplete)

		return nil
	})

	t.NoError(err)
	t.mockOrchestrator.AssertExpectations(t.T())
	ss.AssertExpectations(t.T())
}

func (t *FirstPreStartCompleteTestSuite) TestFirstPreStartCompleteStreamInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	interceptor := NewFirstPreStartCompleteInterceptor(t.mockOrchestrator)
	err := interceptor.FirstPreStartCompleteStreamInterceptor(srv, ss, info, func(srv any, stream grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
	t.mockOrchestrator.AssertExpectations(t.T())
	ss.AssertExpectations(t.T())
}
