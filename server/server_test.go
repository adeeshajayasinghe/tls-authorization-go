package server

import (
	"context"
	"go-authorization/auth"
	"go-authorization/config"
	"go-authorization/log"
	spec "go-authorization/spec"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	for scenarios, fn := range map[string]func(
		t *testing.T,
		rootClient spec.LogClient,
		readOnlyClient spec.LogClient,
		nobodyClient spec.LogClient,
		config *Config,
	){
		"produce/consume a message to/from the log succeeeds": testProduceConsume,
		"readonly write fails and consume succeeds": testReadOnly,
		"unauthorized fails": testUnauthorized,

	}{
		t.Run(scenarios, func(t *testing.T) {
			rootClient, readOnlyClient, nobodyClient, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, readOnlyClient, nobodyClient, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (
	rootClient spec.LogClient,
	readOnlyClient spec.LogClient,
	nobodyClient spec.LogClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()

	// Defining port as 0 will automatically assign us a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func (crtPath, keyPath string) (
		*grpc.ClientConn,
		spec.LogClient,
		[]grpc.DialOption,
	){
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile: keyPath,
			CAFile: config.CAFile,
			Server: false,
		})
		require.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		conn, err := grpc.Dial(l.Addr().String(), opts...)
		require.NoError(t, err)

		client := spec.NewLogClient(conn)
		return conn, client, opts
	}

	// Super user client who is permitted to produce and consume
	var rootConn *grpc.ClientConn
	rootConn, rootClient, _ = newClient(
		config.RootClientCertFile,
 		config.RootClientKeyFile,
	) 

	// Client who is permitted to consume and but not to produce
	var readOnlyConn *grpc.ClientConn
	readOnlyConn, readOnlyClient, _ = newClient(
		config.ReadOnlyClientCertFile,
		config.ReadOnlyClientKeyFile,
	)
	
	// Nobody client who isn't permitted to do anything
	var nobodyConn *grpc.ClientConn
	nobodyConn, nobodyClient, _ = newClient(
		config.NobodyClientCertFile,
 		config.NobodyClientKeyFile,
	)	
	
	severTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.ServerCertFile,
		KeyFile: config.ServerKeyFile,
		CAFile: config.CAFile,
		ServerAddress: l.Addr().String(),
		Server: true,
	})
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(severTLSConfig)

	clog := log.NewLog()

	// Use our authorizor
	authorizer := auth.New(config.ACLModelFile, config.ACLPolicyFile)

	cfg = &Config{
		CommitLog: clog,
		Authorizer: authorizer,
	}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func ()  {
		server.Serve(l)	
	}()
	
	return rootClient, readOnlyClient, nobodyClient, cfg, func() {
		server.Stop()
		rootConn.Close()
		readOnlyConn.Close()
		nobodyConn.Close()
		l.Close()
	}
}
// Use root client to perform operations
func testProduceConsume(t *testing.T, client, _,  _ spec.LogClient, cfg *Config) {
	ctx := context.Background()

	want := &spec.Record{
		Value: []byte("hello world"),
	}
	produce, err := client.Produce(
		ctx, 
		&spec.ProduceRequest{
			Record: want,
		},
	)
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &spec.ConsumeRequest{
		Offset: produce.Offset,
	})
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

// Use nobody client to perform operations
func testReadOnly(
	t *testing.T,
	root spec.LogClient,
	client spec.LogClient,
	_ spec.LogClient,
	config *Config,
) {
	ctx := context.Background()

	// Read-only user writes
	produce, err := client.Produce(ctx,
		&spec.ProduceRequest{
			Record: &spec.Record{
				Value: []byte("hello world"),
			},
		},
	)
	if produce != nil {
		t.Fatal("produce response should be nil")
	}
	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}

	want := &spec.Record{
		Value: []byte("hello world"),
	}

	// Root user writes
	produce, err = root.Produce(ctx,
		&spec.ProduceRequest{
			Record: want,
		},
	)
	require.NoError(t, err)

	// Read-only user reads which root just wrote
	consume, err := client.Consume(ctx,
		&spec.ConsumeRequest{
			Offset: produce.Offset,
		},
	)
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

// Use nobody client to perform operations
func testUnauthorized(
	t *testing.T,
	_,
	_,
	client spec.LogClient,
	config *Config,
) {
	ctx := context.Background()
	produce, err := client.Produce(ctx,
		&spec.ProduceRequest{
			Record: &spec.Record{
				Value: []byte("hello world"),
			},
		},
	)
	if produce != nil {
		t.Fatal("produce response should be nil")
	}
	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}
	consume, err := client.Consume(ctx,
		&spec.ConsumeRequest{
			Offset: 0,
		},
	)
	if consume != nil {
		t.Fatal("consume response should be nil")
	}

	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}
}
