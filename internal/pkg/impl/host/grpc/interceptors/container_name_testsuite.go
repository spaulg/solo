package interceptors

import (
	"context"
	"errors"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpc_mock "github.com/spaulg/solo/test/mocks/grpc"
	"github.com/spaulg/solo/test/mocks/host/container"
)

type ContainerNameTestSuite struct {
	suite.Suite

	mockOrchestrator *container.MockOrchestrator
}

func (t *ContainerNameTestSuite) SetupTest() {
	t.mockOrchestrator = &container.MockOrchestrator{}
}

func (t *ContainerNameTestSuite) TestSuccessfulContainerNameUnaryInterceptor() {
	md := metadata.MD{}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	expectedContainerName := "test-container"
	t.mockOrchestrator.On("ResolveContainerNameFromMetadata", md).Return(&expectedContainerName, nil)

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	result, err := interceptor.ContainerNameUnaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
		containerName, ok := ctx.Value(ContainerName(ContainerNameContextValueName)).(string)

		t.True(ok)
		t.Equal(expectedContainerName, containerName)

		return expectedResult, nil
	})

	t.Equal(expectedResult, result)
	t.NoError(err)
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *ContainerNameTestSuite) TestContainerNameUnaryInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	_, err := interceptor.ContainerNameUnaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *ContainerNameTestSuite) TestContainerNameUnaryInterceptorWithResolveFailure() {
	md := metadata.MD{}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	unexpectedResult := new(interface{})

	t.mockOrchestrator.On("ResolveContainerNameFromMetadata", md).Return(nil, errors.New("mock mockOrchestrator error"))

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	result, err := interceptor.ContainerNameUnaryInterceptor(ctx, req, info, func(ctx context.Context, req any) (any, error) {
		return unexpectedResult, nil
	})

	t.Nil(result)
	t.ErrorContains(err, "mock mockOrchestrator error")
	t.ErrorContains(err, "failed to resolve container name from metadata")
	t.mockOrchestrator.AssertExpectations(t.T())
}

func (t *ContainerNameTestSuite) TestSuccessfulContainerNameStreamInterceptor() {
	md := metadata.MD{}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	expectedContainerName := "test-container"
	t.mockOrchestrator.On("ResolveContainerNameFromMetadata", md).Return(&expectedContainerName, nil)

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	err := interceptor.ContainerNameStreamInterceptor(srv, ss, info, func(srv any, stream grpc.ServerStream) error {
		containerName, ok := stream.Context().Value(ContainerName(ContainerNameContextValueName)).(string)

		t.True(ok)
		t.Equal(expectedContainerName, containerName)

		return nil
	})

	t.NoError(err)
	t.mockOrchestrator.AssertExpectations(t.T())
	ss.AssertExpectations(t.T())
}

func (t *ContainerNameTestSuite) TestContainerNameStreamInterceptorWithMetadataLoadError() {
	ctx := context.Background()

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	err := interceptor.ContainerNameStreamInterceptor(srv, ss, info, func(srv any, stream grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to load metadata from incoming context")
	t.mockOrchestrator.AssertExpectations(t.T())
	ss.AssertExpectations(t.T())
}

func (t *ContainerNameTestSuite) TestContainerNameStreamInterceptorWithResolveFailure() {
	md := metadata.MD{}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(ctx)

	t.mockOrchestrator.On("ResolveContainerNameFromMetadata", md).Return(nil, errors.New("mock mockOrchestrator error"))

	interceptor := NewContainerNameInterceptor(t.mockOrchestrator)
	err := interceptor.ContainerNameStreamInterceptor(srv, ss, info, func(srv any, stream grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "mock mockOrchestrator error")
	t.ErrorContains(err, "failed to resolve container name from metadata")
	t.mockOrchestrator.AssertExpectations(t.T())
	ss.AssertExpectations(t.T())
}
