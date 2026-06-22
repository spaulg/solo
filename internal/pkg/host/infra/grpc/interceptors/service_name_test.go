package interceptors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	grpc_mock "github.com/spaulg/solo/test/mocks/grpc"
)

func TestServiceNameTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceNameTestSuite))
}

type ServiceNameTestSuite struct {
	suite.Suite

	ctx context.Context
	p   *peer.Peer
}

func (t *ServiceNameTestSuite) SetupTest() {
	t.p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{
						Subject: pkix.Name{
							CommonName: "service:test_service_name",
						},
					},
				},
			},
		},
	}

	t.ctx = peer.NewContext(context.Background(), t.p)
}

func (t *ServiceNameTestSuite) TestSuccessfulServiceNameUnaryInterceptor() {
	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewServiceNameInterceptor()
	result, err := interceptor.ServiceNameUnaryInterceptor(t.ctx, req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.Equal(expectedResult, result)
	t.NoError(err)
}

func (t *ServiceNameTestSuite) TestServiceNameUnaryInterceptorWithFailureToFindPeer() {
	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	interceptor := NewServiceNameInterceptor()
	_, err := interceptor.ServiceNameUnaryInterceptor(context.Background(), req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "unable to find peer transport credentials")
}

func (t *ServiceNameTestSuite) TestServiceNameUnaryInterceptorWithFailureToCastToTLSInfo() {
	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	t.p.AuthInfo = nil

	interceptor := NewServiceNameInterceptor()
	_, err := interceptor.ServiceNameUnaryInterceptor(t.ctx, req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "unable to cast transport credentials to TLSInfo")
}

func (t *ServiceNameTestSuite) TestServiceNameUnaryInterceptorWithFailureToFindPeerCertificates() {
	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	tlsInfo, ok := t.p.AuthInfo.(credentials.TLSInfo)
	t.True(ok)

	tlsInfo.State.PeerCertificates = []*x509.Certificate{}
	t.p.AuthInfo = tlsInfo

	interceptor := NewServiceNameInterceptor()
	_, err := interceptor.ServiceNameUnaryInterceptor(t.ctx, req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "missing peer certificate")
}

func (t *ServiceNameTestSuite) TestServiceNameUnaryInterceptorWithInvalidCommonName() {
	info := &grpc.UnaryServerInfo{}
	req := new(interface{})
	expectedResult := new(interface{})

	t.p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{
						Subject: pkix.Name{
							CommonName: "invalid_common_name_format",
						},
					},
				},
			},
		},
	}

	t.ctx = peer.NewContext(context.Background(), t.p)

	interceptor := NewServiceNameInterceptor()
	_, err := interceptor.ServiceNameUnaryInterceptor(t.ctx, req, info, func(_ context.Context, _ any) (any, error) {
		return expectedResult, nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "invalid subject common name")
}

func (t *ServiceNameTestSuite) TestSuccessfulServiceNameStreamInterceptor() {
	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(t.ctx)

	interceptor := NewServiceNameInterceptor()
	err := interceptor.ServiceNameStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.NoError(err)
	ss.AssertExpectations(t.T())
}

func (t *ServiceNameTestSuite) TestServiceNameStreamInterceptorWithFailureToFindPeer() {
	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(context.Background())

	interceptor := NewServiceNameInterceptor()
	err := interceptor.ServiceNameStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "unable to find peer transport credentials")
	ss.AssertExpectations(t.T())
}

func (t *ServiceNameTestSuite) TestServiceNameStreamInterceptorWithFailureToCastToTLSInfo() {
	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(t.ctx)

	t.p.AuthInfo = nil

	interceptor := NewServiceNameInterceptor()
	err := interceptor.ServiceNameStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "unable to cast transport credentials to TLSInfo")
	ss.AssertExpectations(t.T())
}

func (t *ServiceNameTestSuite) TestServiceNameStreamInterceptorWithFailureToFindPeerCertificates() {
	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(t.ctx)

	tlsInfo := t.p.AuthInfo.(credentials.TLSInfo)
	tlsInfo.State.PeerCertificates = []*x509.Certificate{}
	t.p.AuthInfo = tlsInfo

	interceptor := NewServiceNameInterceptor()
	err := interceptor.ServiceNameStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "missing peer certificate")
	ss.AssertExpectations(t.T())
}

func (t *ServiceNameTestSuite) TestServiceNameStreamInterceptorWithInvalidCommonName() {
	info := &grpc.StreamServerInfo{}
	srv := new(interface{})

	t.p = &peer.Peer{
		AuthInfo: credentials.TLSInfo{
			State: tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{
						Subject: pkix.Name{
							CommonName: "invalid_common_name_format",
						},
					},
				},
			},
		},
	}

	t.ctx = peer.NewContext(context.Background(), t.p)

	ss := &grpc_mock.MockServerStream{}
	ss.On("Context").Return(t.ctx)

	interceptor := NewServiceNameInterceptor()
	err := interceptor.ServiceNameStreamInterceptor(srv, ss, info, func(_ any, _ grpc.ServerStream) error {
		return nil
	})

	t.ErrorContains(err, "failed to find service name")
	t.ErrorContains(err, "invalid subject common name")
	ss.AssertExpectations(t.T())
}
