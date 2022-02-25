package grpc

import (
	"context"
	"crypto/rsa"
	"log"
	"net"
	"optisam-backend/common/optisam/logger"
	mw "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/report-service/pkg/api/v1"
	"os"
	"os/signal"

	"github.com/open-policy-agent/opa/rego"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

// RunServer runs gRPC service to publish Auth service
func RunServer(ctx context.Context, v1API v1.ReportServiceServer, port string, verifyKey *rsa.PublicKey, p *rego.PreparedEvalQuery, apiKey string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// gRPC server statup options
	opts := mw.Chained(logger.Log, verifyKey, p, apiKey)
	opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	// add middleware
	// opts = grpc_middleware.AddLogging(logger.Log, opts)
	// register service
	server := grpc.NewServer(opts...)
	v1.RegisterReportServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Println("shutting down gRPC server...")

			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
