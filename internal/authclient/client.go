package authclient

import (
	authv1 "auth-service/gen/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Auth  authv1.AuthServiceClient
	close func() error
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		Auth:  authv1.NewAuthServiceClient(conn),
		close: conn.Close,
	}, nil
}

func (c *Client) Close() error {
	if c.close == nil {
		return nil
	}
	return c.close()
}
