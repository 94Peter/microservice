package grpc_tool

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/94peter/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Connection interface {
	grpc.ClientConnInterface
	Close() error
	IsValid() bool
	WaitUntilReady() bool
}

var grpcSchemaRegex = regexp.MustCompile(`^grpc(s)?://`)

var kacp = keepalive.ClientParameters{
	Time:                30 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             30 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

func NewConnection(address string) (Connection, error) {
	var conn *grpc.ClientConn
	var err error
	hasPrefix := grpcSchemaRegex.MatchString(address)
	var schema string
	var host string
	if hasPrefix {
		u, err := url.Parse(address)
		if err != nil {
			return nil, fmt.Errorf("address [%s] error: %s", address, err.Error())
		}
		schema = u.Scheme
		host = u.Host
	} else {
		schema = "grpc"
		host = address
	}

	if schema == "grpcs" {
		conn, err = grpc.NewClient(host,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})),
			grpc.WithKeepaliveParams(kacp),
		)
	} else {
		conn, err = grpc.NewClient(host,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(kacp),
		)
	}
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

func NewAutoReconn(address string) *AutoReConn {
	return &AutoReConn{
		address:   address,
		Ready:     make(chan bool),
		Done:      make(chan bool),
		Reconnect: make(chan bool),
		Error:     make(chan error),
	}
}

type AutoReConnInter interface {
	SetLog(l log.Logger)
}

type AutoReConn struct {
	Connection

	address string

	Ready     chan bool
	Done      chan bool
	Reconnect chan bool
	Error     chan error

	log log.Logger
}

type GetGrpcFunc func(myGrpc Connection) error

func (my *AutoReConn) Connect() (Connection, error) {
	return NewConnection(my.address)
}

func (my *AutoReConn) Close() error {
	my.Done <- true
	return my.Connection.Close()
}

func (my *AutoReConn) IsValid() bool {
	if my.Connection == nil {
		return false
	}
	return my.Connection.IsValid()
}

func (my *AutoReConn) Process(f GetGrpcFunc) {
	var err error
	isFirst := true
	for {
		if !isFirst {
			time.Sleep(10 * time.Second)
		}
		isFirst = false
		my.Connection, err = my.Connect()
		if err != nil {
			my.printErr(err)
			continue
		}
		if err = f(my.Connection); err != nil {
			my.printErr(err)
			continue
		}
		break
	}
}

func (my *AutoReConn) printErr(err error) {
	if my.log == nil {
		return
	}
	my.log.Warn(err.Error())
}

func (my *AutoReConn) SetLog(l log.Logger) {
	my.log = l
}

func (my *AutoReConn) Start(f GetGrpcFunc) {
	go my.Process(f)
	for {
		select {
		case <-my.Ready:
			continue
		case <-my.Reconnect:
			if !my.WaitUntilReady() {
				continue
			}
		case <-my.Done:
			return
		}
	}
}
