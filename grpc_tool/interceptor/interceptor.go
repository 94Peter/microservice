package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type Interceptor interface {
	StreamServerInterceptor() grpc.StreamServerInterceptor
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
}

func NewSimpleInterceptor(
	stream grpc.StreamServerInterceptor,
	unary grpc.UnaryServerInterceptor) Interceptor {
	return &simpleInterceptor{
		stream: stream,
		unary:  unary,
	}
}

type simpleInterceptor struct {
	stream grpc.StreamServerInterceptor
	unary  grpc.UnaryServerInterceptor
}

func (i *simpleInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return i.stream
}

func (i *simpleInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return i.unary
}

func IsReflectMethod(m string) bool {
	return m == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"
}

func NewServerStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &serverStream{
		ServerStream: stream,
		ctx:          stream.Context(),
	}
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}
