package grpcclient

import (
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"google.golang.org/grpc"
)

type Client struct {
	AuthClient ssov1.AuthClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		AuthClient: ssov1.NewAuthClient(conn),
	}
}
