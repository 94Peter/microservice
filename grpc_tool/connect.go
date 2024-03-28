package grpc_tool

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Connection interface {
	grpc.ClientConnInterface
	Close() error
	IsValid() bool
	WaitUntilReady() bool
}

func NewConnection(ctx context.Context, address string) (Connection, error) {
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("address [%s] error: %s", address, err.Error())
	}
	return &myGrpcImpl{
		ClientConn: conn,
	}, nil
}

type myGrpcImpl struct {
	*grpc.ClientConn
}

func (my *myGrpcImpl) Close() error {
	return my.ClientConn.Close()
}

func (my *myGrpcImpl) IsValid() bool {
	if my.ClientConn == nil {
		return false
	}
	switch my.ClientConn.GetState() {
	case connectivity.Ready:
		return true
	case connectivity.Idle:
		return false
	default:
		return false
	}
}

func (my *myGrpcImpl) WaitUntilReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) //define how long you want to wait for connection to be restored before giving up
	defer cancel()
	return my.WaitForStateChange(ctx, connectivity.Ready)
}

func NewAutoReconn(address string, timeout time.Duration) *AutoReConn {
	return &AutoReConn{
		address:   address,
		timeout:   timeout,
		Ready:     make(chan bool),
		Done:      make(chan bool),
		Reconnect: make(chan bool),
	}
}

type AutoReConn struct {
	Connection

	address string
	timeout time.Duration

	Ready     chan bool
	Done      chan bool
	Reconnect chan bool
}

type GetGrpcFunc func(myGrpc Connection) error

func (my *AutoReConn) Connect(ctx context.Context) (Connection, error) {
	return NewConnection(ctx, my.address)
}

func (my *AutoReConn) IsValid() bool {
	if my.Connection == nil {
		return false
	}
	return my.Connection.IsValid()
}

func (my *AutoReConn) Process(f GetGrpcFunc) {
	var err error
	basicCtx := context.Background()
	var ctx context.Context
	var cancel context.CancelFunc
	for {
		if cancel != nil {
			cancel()
		}
		if my.timeout > 0 {
			ctx, cancel = context.WithTimeout(basicCtx, my.timeout)
		} else {
			ctx = basicCtx
		}

		my.Connection, err = my.Connect(ctx)
		if err != nil {
			continue
		}
		if err = f(my.Connection); err != nil {
			continue
		}
		break
	}
	if cancel != nil {
		cancel()
	}
}
