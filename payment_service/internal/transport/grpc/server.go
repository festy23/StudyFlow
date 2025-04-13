package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"payment_service/internal/service"
	paymentv1 "payment_service/proto"
)

type Server struct {
	server *grpc.Server
	port   int
}

func NewServer(svc *service.PaymentService, cfg config.GRPCConfig) *Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)

	h := newHandler(svc)
	paymentv1.RegisterPaymentServiceServer(srv, h)

	return &Server{
		server: srv,
		port:   cfg.Port,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
