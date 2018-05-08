package grpcmon_test

import (
	"log"
	"net"

	"github.com/Bo0mer/grpcmon"
	bpb "github.com/Bo0mer/grpcmon/testdata/backend"
	pb "github.com/Bo0mer/grpcmon/testdata/frontend"
	"google.golang.org/grpc"
)

func Example() {
	// Create gRPC metrics with selected options and register with monitoring
	// sytem.
	clientMetrics := &grpcmon.Metrics{
	// ...
	}
	// Instrument gRPC client(s).
	backendConn, err := grpc.Dial(backendAddr, grpcmon.DialOption(clientMetrics))
	if err != nil {
		log.Fatal(err)
	}

	serverMetrics := &grpcmon.Metrics{
	// ...
	}
	// Instrument gRPC server and, optionally, initialize server metrics.
	srv := grpc.NewServer(grpcmon.ServerOption(serverMetrics))
	pb.RegisterFrontendServer(srv, &Server{
		backend: bpb.NewBackendClient(backendConn),
	})
	// Listen and serve.
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(srv.Serve(lis))
}
