package transport

import (
	"context"
	"io"

	"github.com/hyperleex/zenmcp/protocol"
)

type Transport interface {
	Accept(ctx context.Context) (Connection, error)
	Close() error
}

type Connection interface {
	Codec() protocol.Codec
	Context() context.Context
	Close() error
}

type Client interface {
	Connect(ctx context.Context) (Connection, error)
	Close() error
}

type ReadWriteCloser interface {
	io.ReadWriteCloser
}

type connImpl struct {
	codec protocol.Codec
	ctx   context.Context
}

func NewConnection(ctx context.Context, codec protocol.Codec) Connection {
	return &connImpl{
		codec: codec,
		ctx:   ctx,
	}
}

func (c *connImpl) Codec() protocol.Codec {
	return c.codec
}

func (c *connImpl) Context() context.Context {
	return c.ctx
}

func (c *connImpl) Close() error {
	return c.codec.Close()
}