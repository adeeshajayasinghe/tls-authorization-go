package server

import (
	"context"
	spec "go-authorization/spec"

	"google.golang.org/grpc"
)


type Config struct {
	CommitLog CommitLog
}

type CommitLog interface {
	Produce(req *spec.ProduceRequest) (uint64, error)
	Consume(req *spec.ConsumeRequest) (*spec.Record, error)
}

var _ spec.LogServer = (*grpcServer)(nil)

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
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
	offset, err := s.CommitLog.Produce(req)
	if err != nil {
		return nil, err
	}

	return &spec.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *spec.ConsumeRequest) (*spec.ConsumeResponse, error) {
	record, err := s.CommitLog.Consume(req)
	if err != nil {
		return nil, err
	}

	return &spec.ConsumeResponse{Record: record}, nil
}




