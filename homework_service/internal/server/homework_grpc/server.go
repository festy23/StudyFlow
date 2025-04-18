package homework_grpc

import (
	"net"

	"google.golang.org/grpc"
	v1 "homework_service/pkg/api"
)

type Server struct {
	server *grpc.Server
}

func NewServer(handler *HomeworkHandler, interceptor grpc.UnaryServerInterceptor) *Server {
	srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	v1.RegisterHomeworkServiceServer(srv, handler)
	return &Server{
		server: srv,
	}
}

func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
