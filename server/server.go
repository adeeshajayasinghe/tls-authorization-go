package server

import (
	"context"
	spec "go-authorization/spec"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)


type Config struct {
	CommitLog CommitLog
	Authorizer Authorizer
}

const (
	objectWildcard = "*"
	produceAction = "produce"
	consumeAction = "consume"
)


type CommitLog interface {
	Produce(req *spec.ProduceRequest) (uint64, error)
	Consume(req *spec.ConsumeRequest) (*spec.Record, error)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
}

var _ spec.LogServer = (*grpcServer)(nil)

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// Middleware that intercept and modify the execution of the RPC call
	opts = append(opts, grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(authenticate),
			)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(authenticate),
	)))

	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	spec.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}

type grpcServer struct {
	spec.UnimplementedLogServer
	*Config
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *spec.ProduceRequest) (*spec.ProduceResponse, error) {
	// Checks whether client is authorize to produce
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		produceAction,
	); err != nil {
		return nil, err
	}

	offset, err := s.CommitLog.Produce(req)
	if err != nil {
		return nil, err
	}

	return &spec.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *spec.ConsumeRequest) (*spec.ConsumeResponse, error) {
	// Checks whether client is authorize to consume
	if err := s.Authorizer.Authorize(
		subject(ctx),
		objectWildcard,
		consumeAction,
	); err != nil {
		return nil, err
	}
	record, err := s.CommitLog.Consume(req)
	if err != nil {
		return nil, err
	}

	return &spec.ConsumeResponse{Record: record}, nil
}

// Filter the subject information from the client certificate
// and write it to the RPC context
func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	ctx = context.WithValue(ctx, subjectContextKey{}, subject)
	return ctx, nil
}

// Filter the subject information from the RPC context
func subject(ctx context.Context) string {	
	return ctx.Value(subjectContextKey{}).(string)
}

type subjectContextKey struct{}




