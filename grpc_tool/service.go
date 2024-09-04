package grpc_tool

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      4 * time.Hour,    // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 10 * time.Second, // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

func RunGrpcServ(ctx context.Context, cfg *GrpcConfig) error {
	if cfg.registerServiceFunc == nil {
		return fmt.Errorf("registerServiceFunc must not be nil")
	}
	port := ":" + strconv.Itoa(cfg.Port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	var serv *grpc.Server
	if len(cfg.interceptors) > 0 {
		var streamInterceptors []grpc.StreamServerInterceptor
		var unaryInterceptors []grpc.UnaryServerInterceptor
		for _, i := range cfg.interceptors {
			streamInterceptors = append(streamInterceptors, i.StreamServerInterceptor())
			unaryInterceptors = append(unaryInterceptors, i.UnaryServerInterceptor())
		}
		serv = grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		)
	} else {
		serv = grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	}
	if cfg.ReflectService {
		reflection.Register(serv)
	}
	cfg.registerServiceFunc(serv)
	var grpcWait sync.WaitGroup
	grpcWait.Add(1)
	go func(s *grpc.Server, lis net.Listener, l Log) {
		for {
			l.Infof("app gRPC server is running [%s].", lis.Addr())
			if err := s.Serve(lis); err != nil {
				switch err {
				case grpc.ErrServerStopped:
					grpcWait.Done()
					return
				default:
					l.Fatalf("failed to serve: %v", err)
				}
			}
		}
	}(serv, lis, cfg.Logger)
	<-ctx.Done()
	serv.Stop()
	grpcWait.Wait()
	return nil
}
