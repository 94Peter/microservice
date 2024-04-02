package interceptor

import (
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
