package grpc

type MockBidiStreamingServer[Req any, Res any] struct {
	MockServerStream
}

func (m *MockBidiStreamingServer[Req, Res]) Recv() (*Req, error) {
	args := m.Called()

	if req, ok := args.Get(0).(*Req); ok {
		return req, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockBidiStreamingServer[Req, Res]) Send(res *Res) error {
	args := m.Called(res)
	return args.Error(0)
}
