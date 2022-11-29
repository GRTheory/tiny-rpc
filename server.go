package tinyrpc

import (
	"log"
	"net"
	"net/rpc"

	"github.com/GRTheory/tiny-rpc/codec"
	"github.com/GRTheory/tiny-rpc/serializer"
)

type Server struct {
	*rpc.Server
	serializer.Serializer
}

func NewServer(opts ...Option) *Server {
	options := options{
		serializer: serializer.Proto,
	}
	for _, option := range opts {
		option(&options)
	}

	return &Server{&rpc.Server{}, options.serializer}
}

func (s *Server) Register(rcvr interface{}) error{
	return s.Server.Register(rcvr)
}

func (s *Server) RegisterName(name string, rcvr interface{}) error{
	return s.Server.RegisterName(name, rcvr)
}

func (s *Server) Serve(lis net.Listener){
	log.Printf("tinyrpc started on: %s", lis.Addr().String())
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn, s.Serializer))
	}
}
