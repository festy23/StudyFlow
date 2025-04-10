package grpc

import (
	"net"

	v1 "homework_service/pkg/api"

	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	config Config
}

type Config struct {
	Address string
}

func NewServer(config Config, handler *HomeworkHandler) *Server {
	srv := grpc.NewServer()
	v1.RegisterHomeworkServiceServer(srv, handler)

	return &Server{
		server: srv,
		config: config,
	}
}

func (s *Server) Serve(lis net.Listener) error {
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
