package grpcprom_test

import (
	"log"
	"net"

	"github.com/Bo0mer/grpcprom"
	bpb "github.com/Bo0mer/grpcprom/testdata/backend"
	pb "github.com/Bo0mer/grpcprom/testdata/frontend"
	"google.golang.org/grpc"
)

func Example() {
	// Create gRPC metrics with selected options and register with monitoring
	// sytem.
	clientMetrics := &grpcprom.Metrics{
	// ...
	}
	// Instrument gRPC client(s).
	backendConn, err := grpc.Dial(backendAddr, grpcprom.DialOption(clientMetrics))
	if err != nil {
		log.Fatal(err)
	}

	serverMetrics := &grpcprom.Metrics{
	// ...
	}
	// Instrument gRPC server and, optionally, initialize server metrics.
	srv := grpc.NewServer(grpcprom.ServerOption(serverMetrics))
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
