// Package grpcmon provides monitoring instrumentation for gRPC clients and
// servers.
//
// The following metrics are provided:
//
//  grpc_client_connections_open [gauge] Number of gRPC client connections open.
//  grpc_client_connections_total [counter] Total number of gRPC client connections opened.
//  grpc_client_requests_pending{service,method} [gauge] Number of gRPC client requests pending.
//  grpc_client_requests_total{service,method,code} [counter] Total number of gRPC client requests completed.
//  grpc_client_latency_seconds{service,method,code} [histogram] Latency of gRPC client requests.
//  grpc_client_recv_bytes{service,method,frame} [histogram] Bytes received in gRPC client responses.
//  grpc_client_sent_bytes{service,method,frame} [histogram] Bytes sent in gRPC client requests.
//
//  grpc_server_connections_open [gauge] Number of gRPC server connections open.
//  grpc_server_connections_total [counter] Total number of gRPC server connections opened.
//  grpc_server_requests_pending{service,method} [gauge] Number of gRPC server requests pending.
//  grpc_server_requests_total{service,method,code} [counter] Total number of gRPC server requests completed.
//  grpc_server_latency_seconds{service,method,code} [histogram] Latency of gRPC server requests.
//  grpc_server_recv_bytes{service,method,frame} [histogram] Bytes received in gRPC server requests.
//  grpc_server_sent_bytes{service,method,frame} [histogram] Bytes sent in gRPC server responses.
package grpcmon // import "github.com/Bo0mer/grpcmon"

import (
	"context"
	"strings"
	"time"

	metrics "github.com/go-kit/kit/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

const (
	header  = "header"
	payload = "payload"
	trailer = "trailer"
)

// DefaultLatencyBuckets provides convenient default latency histogram buckets.
var DefaultLatencyBuckets = []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// DefaultBytesBuckets provides convenient default bytes histogram buckets.
var DefaultBytesBuckets = []float64{0, 32, 64, 128, 256, 512, 1024, 2048, 8192, 32768, 131072, 524288}

// DialOption returns a gRPC DialOption that instruments metrics
// for the client connection.
func DialOption(metrics *Metrics) grpc.DialOption {
	return grpc.WithStatsHandler(&handler{client: metrics})
}

// ServerOption returns a gRPC ServerOption that instruments metrics
// for the server.
func ServerOption(metrics *Metrics) grpc.ServerOption {
	return grpc.StatsHandler(&handler{server: metrics})
}

// Metrics tracks gRPC metrics.
type Metrics struct {
	_ struct{}

	ConnsOpen   metrics.Gauge
	ConnsTotal  metrics.Counter
	ReqsPending metrics.Gauge
	ReqsTotal   metrics.Counter
	Latency     metrics.Histogram
	BytesSent   metrics.Histogram
	BytesRecv   metrics.Histogram
}

var rpcInfoKey = "rpc-tag"

type rpcInfo struct {
	server string
	method string
	begin  time.Time
}

// handler implements the stats.Handler interface.
type handler struct {
	client *Metrics
	server *Metrics
}

// TagRPC implements the stats.Handler interface.
func (*handler) TagRPC(ctx context.Context, v *stats.RPCTagInfo) context.Context {
	server, method := splitFullMethodName(v.FullMethodName)
	return context.WithValue(ctx, &rpcInfoKey, &rpcInfo{
		server: server,
		method: method,
	})
}

func splitFullMethodName(s string) (server, method string) {
	s = strings.TrimPrefix(s, "/")
	i := strings.Index(s, "/")
	if i < 0 {
		return "unknown", "unknown"
	}
	return s[:i], s[i+1:]
}

// HandleRPC implements the stats.Handler interface.
func (h *handler) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	v, ok := ctx.Value(&rpcInfoKey).(*rpcInfo)
	if !ok {
		return
	}
	m := h.server
	if stat.IsClient() {
		m = h.client
	}
	switch s := stat.(type) {
	case *stats.Begin:
		v.begin = s.BeginTime
		m.ReqsPending.With("service", v.server, "method", v.method).Add(1)
	case *stats.End:
		code := status.Code(s.Error).String()
		if m.Latency != nil {
			m.Latency.With("service", v.server, "method", v.method, "code", code).Observe(time.Since(v.begin).Seconds())
		}
		m.ReqsTotal.With("service", v.server, "method", v.method, "code", code).Add(1)
		m.ReqsPending.With("service", v.server, "method", v.method).Add(-1)
	case *stats.InHeader:
		if m.BytesRecv != nil {
			m.BytesRecv.With("service", v.server, "method", v.method, "frame", header).Observe(float64(s.WireLength))
		}
	case *stats.InPayload:
		if m.BytesRecv != nil {
			m.BytesRecv.With("service", v.server, "method", v.method, "frame", payload).Observe(float64(s.WireLength))
		}
	case *stats.InTrailer:
		if m.BytesRecv != nil {
			m.BytesRecv.With("service", v.server, "method", v.method, "frame", trailer).Observe(float64(s.WireLength))
		}
	case *stats.OutHeader:
		if m.BytesSent != nil {
			m.BytesSent.With(v.server, v.method, header).Observe(0) // TODO ???
		}
	case *stats.OutPayload:
		if m.BytesSent != nil {
			m.BytesSent.With("service", v.server, "method", v.method, "frame", payload).Observe(float64(s.WireLength))
		}
	case *stats.OutTrailer:
		if m.BytesSent != nil {
			m.BytesSent.With("service", v.server, "method", v.method, "frame", trailer).Observe(float64(s.WireLength))
		}
	}
}

// TagConn implements the stats.Handler interface.
func (h *handler) TagConn(ctx context.Context, v *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn implements the stats.Handler interface.
func (h *handler) HandleConn(ctx context.Context, stat stats.ConnStats) {
	m := h.server
	if stat.IsClient() {
		m = h.client
	}
	switch stat.(type) {
	case *stats.ConnBegin:
		m.ConnsOpen.Add(1)
		m.ConnsTotal.Add(1)
	case *stats.ConnEnd:
		m.ConnsOpen.Add(-1)
	}
}
