package ceph

import "errors"

type Client struct {
	Session       *Session
	MaxIterations uint
	Logger        *Adapter
}

var ErrMaxIterationsExceeded = errors.New("max recursive iterations exceeded")

func New(server Server) (client *Client, err error) {

	client = &Client{MaxIterations: 30}
	if client.Logger == nil {
		client.Logger = NewLogger()
	}

	client.Session, err = NewSession(server)

	if err != nil {
		return nil, err
	}

	return client, nil
}
