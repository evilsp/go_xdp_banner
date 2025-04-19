package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type HttpServer struct {
	cancel context.CancelFunc
	mux    *runtime.ServeMux
	conn   *grpc.ClientConn
	server *http.Server
}

func NewHttpServer(grpcListenAddr string, service server.Services) HttpServer {
	grpcEndpoint, err := localGrpcEndpoint(grpcListenAddr)
	if err != nil {
		log.Fatal("failed to get local grpc endpoint", log.ErrorField(err))
	}

	creds, err := NewCredits()
	if err != nil {
		log.Fatal("failed to create credentials", log.ErrorField(err))
	}

	conn, err := grpc.NewClient(
		grpcEndpoint,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatal("init Http gateway client failed", log.ErrorField(err))
	}

	// convert Services to HttpServices
	httpService := make(server.HttpServices, len(service))
	for name, s := range service {
		httpService[name] = s
	}

	ctx, cancel := context.WithCancel(context.Background())
	mux := runtime.NewServeMux()
	if err = setupHttpHandler(httpService, ctx, mux, conn); err != nil {
		cancel()
		log.Fatal("setup http gateway failed", log.ErrorField(err))
	}

	return HttpServer{
		cancel: cancel,
		mux:    mux,
		conn:   conn,
	}
}

func (s *HttpServer) Serve(addr string) error {
	log.Info("Starting HTTP gateway on port " + addr)
	s.server = &http.Server{Addr: addr, Handler: s.mux}
	return s.server.ListenAndServe()
}

func (s *HttpServer) Close() {
	s.cancel()
	s.server.Close()
	s.conn.Close()
}

func setupHttpHandler(service server.HttpServices, ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	for n, s := range service {
		if err := s.RegisterHttpService(ctx, mux, conn); err != nil {
			return fmt.Errorf("setup %s service failed: %w", n, err)
		}
	}

	return nil
}

func localGrpcEndpoint(listenAddr string) (string, error) {
	_, port, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort("localhost", port), nil
}
