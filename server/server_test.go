package server

import (
	"context"
	"go-authorization/log"
	spec "go-authorization/spec"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	ctx := context.Background()
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{grpc.WithInsecure()}
	cc, err := grpc.Dial(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	clog := log.NewLog()

	cfg := &Config{
		CommitLog: clog,
	}
	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func ()  {
		server.Serve(l)	
	}()

	client := spec.NewLogClient(cc)
	
	want := &spec.Record{
		Value: []byte("hello world"),
	}

	produce, err := client.Produce(
		ctx,
		&spec.ProduceRequest{Record: want},
	)
	require.NoError(t, err)

	consume, err := client.Consume(
		ctx,
		&spec.ConsumeRequest{Offset: produce.Offset},
	)
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)

}