package server

import (
	"context"
	"go-authorization/config"
	"go-authorization/log"
	spec "go-authorization/spec"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestServer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	ctx := context.Background()
	require.NoError(t, err)

	clientTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CAFile: config.CAFile,
		KeyFile: config.ClientKeyFile,
		CertFile: config.ClientCertFile,
	})
	require.NoError(t, err)

	clientCred := credentials.NewTLS(clientTLSConfig)

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithTransportCredentials(clientCred))
	require.NoError(t, err)

	client := spec.NewLogClient(cc)

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.ServerCertFile,
		KeyFile: config.ServerKeyFile,
		CAFile: config.CAFile,
		ServerAddress: l.Addr().String(),
		Server: true,
	})
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(serverTLSConfig)

	clog := log.NewLog()

	cfg := &Config{
		CommitLog: clog,
	}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func ()  {
		server.Serve(l)	
	}()
	
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