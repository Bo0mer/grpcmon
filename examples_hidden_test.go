package grpcprom_test

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bpb "github.com/Bo0mer/grpcprom/testdata/backend"
	pb "github.com/Bo0mer/grpcprom/testdata/frontend"
)

var addr, backendAddr string

type Server struct {
	backend bpb.BackendClient
}

func (*Server) Query(context.Context, *pb.QueryRequest) (*pb.QueryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Query not implemented")
}
