package interceptors

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/metadata"

	grpc_mocks "github.com/spaulg/solo/test/mocks/grpc"
)

func TestStreamWrapperTestSuite(t *testing.T) {
	suite.Run(t, new(StreamWrapperTestSuite))
}

type StreamWrapperTestSuite struct {
	suite.Suite

	mockStream  *grpc_mocks.MockServerStream
	mockContext *grpc_mocks.MockContext
}

func (t *StreamWrapperTestSuite) SetupTest() {
	t.mockStream = &grpc_mocks.MockServerStream{}
	t.mockContext = &grpc_mocks.MockContext{}
}

func (t *StreamWrapperTestSuite) TestSetHeader() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	md := metadata.MD{"key": []string{"value"}}
	t.mockStream.On("SetHeader", md).Return(nil)

	err := steamWrapper.SetHeader(md)
	t.NoError(err)
}

func (t *StreamWrapperTestSuite) TestSendHeader() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	md := metadata.MD{"key": []string{"value"}}
	t.mockStream.On("SendHeader", md).Return(nil)

	err := steamWrapper.SendHeader(md)
	t.NoError(err)
}

func (t *StreamWrapperTestSuite) TestSetTrailer() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	md := metadata.MD{"key": []string{"value"}}
	t.mockStream.On("SetTrailer", md).Return(nil)

	steamWrapper.SetTrailer(md)
}

func (t *StreamWrapperTestSuite) TestContext() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	t.mockStream.On("Context").Return(t.mockContext)

	context := steamWrapper.Context()
	t.Equal(t.mockContext, context)
}

func (t *StreamWrapperTestSuite) TestSendMsg() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	t.mockStream.On("SendMsg", "message").Return(nil)

	err := steamWrapper.SendMsg("message")
	t.NoError(err)
}

func (t *StreamWrapperTestSuite) TestRecvMsg() {
	steamWrapper := NewServerStreamWrapper(t.mockStream, t.mockContext)

	t.mockStream.On("RecvMsg", "message").Return(nil)

	err := steamWrapper.RecvMsg("message")
	t.NoError(err)
}
