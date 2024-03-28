package grpc_tool

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		)
	} else {
		serv = grpc.NewServer()
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
