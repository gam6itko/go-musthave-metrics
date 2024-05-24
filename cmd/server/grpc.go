package main

import (
	"context"
	grpc2 "github.com/gam6itko/go-musthave-metrics/internal/server/grpc"
	"github.com/gam6itko/go-musthave-metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

func runGRPCServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if Cfg.GRPCAddress == "" {
		Log.Info("gRPC server not started. Address not defined.")
		return
	}

	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc2.RequestLoggingInterceptor),
	)
	proto.RegisterMetricsServer(server, grpc2.NewMetricsServerImpl(MetricStorage))

	go func() {
		<-ctx.Done()

		Log.Info("Shutting down server gracefully")
		server.GracefulStop()
	}()

	Log.Info("Starting gRPC server", zap.String("addr", Cfg.GRPCAddress))
	if err := server.Serve(listen); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()

	Log.Info("gRPC server stopped.")
}
