package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	grpc_mock "github.com/spaulg/solo/test/mocks/grpc"
)

func TestFirstPreStartCompleteTestSuite(t *testing.T) {
	suite.Run(t, new(FirstPreStartCompleteTestSuite))
}

type FirstPreStartCompleteTestSuite struct {
	suite.Suite
}

func (t *FirstPreStartCompleteTestSuite) TestSuccessfulFirstPreStartCompleteUnaryInterceptor() {
	expectedFirstPreStartComplete := "true"
	md := metadata.MD{}
	md.Set(workflowcommon.FirstPreStartContainer.String()+"_complete", expectedFirstPreStartComplete)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewFirstContainerCompleteInterceptor()
	result, err := interceptor.FirstContainerCompleteUnaryInterceptor(ctx, req, info, func(ctx context.Context, _ any) (any, error) {
		firstPreStartComplete, ok := ctx.Value(FirstContainerComplete(workflowcommon.FirstPreStartContainer)).(string)

		t.True(ok)
		t.Equal(expectedFirstPreStartComplete, firstPreStartComplete)

		return expectedResult, nil
	})

	t.Equal(expectedResult, result)
	t.NoError(err)
}

func (t *FirstPreStartCompleteTestSuite) TestFirstPreStartCompleteUnaryInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewFirstContainerCompleteInterceptor()
	_, err := interceptor.FirstContainerCompleteUnaryInterceptor(ctx, req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
}

func (t *FirstPreStartCompleteTestSuite) TestSuccessfulFirstPreStartCompleteStreamInterceptor() {
	expectedFirstPreStartComplete := "true"
	md := metadata.MD{}
	md.Set(workflowcommon.FirstPreStartContainer.String()+"_complete", expectedFirstPreStartComplete)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	interceptor := NewFirstContainerCompleteInterceptor()
	err := interceptor.FirstContainerCompleteStreamInterceptor(srv, ss, info, func(_ any, stream grpc.ServerStream) error {
		firstPreStartComplete, ok := stream.Context().Value(FirstContainerComplete(workflowcommon.FirstPreStartContainer)).(string)

		t.True(ok)
		t.Equal(expectedFirstPreStartComplete, firstPreStartComplete)

		return nil
	})

	t.NoError(err)
	ss.AssertExpectations(t.T())
}

func (t *FirstPreStartCompleteTestSuite) TestFirstPreStartCompleteStreamInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	interceptor := NewFirstContainerCompleteInterceptor()
	err := interceptor.FirstContainerCompleteStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
	ss.AssertExpectations(t.T())
}
