package tinyrpc

import (
	"io"
	"net/rpc"

	"github.com/GRTheory/tiny-rpc/codec"
	"github.com/GRTheory/tiny-rpc/compressor"
	"github.com/GRTheory/tiny-rpc/serializer"
)

type Client struct {
	*rpc.Client
}

// Option provides options for rpc
type Option func(o *options)

type options struct {
	compressType compressor.CompressType
	serializer   serializer.Serializer
}

func WithCompress(c compressor.CompressType) Option {
	return func(o *options) {
		o.compressType = c
	}
}

func WithSerilizer(serializer serializer.Serializer) Option {
	return func(o *options) {
		o.serializer = serializer
	}
}

func NewClient(conn io.ReadWriteCloser, opts ...Option) *Client {
	options := options{
		compressType: compressor.Raw,
		serializer:   serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}
	return &Client{rpc.NewClientWithCodec(
		codec.NewClientCodec(conn, options.compressType, options.serializer),
	)}
}

func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return c.Client.Call(serviceMethod, args, reply)
}

func (c *Client) AsyncCall(serviceMethod string, args interface{}, reply interface{}) chan *rpc.Call {
	return c.Go(serviceMethod, args, reply, nil).Done
}
