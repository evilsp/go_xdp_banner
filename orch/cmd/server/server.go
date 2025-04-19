package server

import (
	"context"
	"fmt"
	"sync"
	"xdp-banner/orch/logic"
	"xdp-banner/orch/server"
	"xdp-banner/orch/service"
	"xdp-banner/pkg/log"
)

func runServer(ctx context.Context, opt *Option, logic *logic.Logic, wg *sync.WaitGroup, errChan chan error) {
	services := service.New(logic)

	grpcServer := server.NewGrpcServer(services)
	httpServer := server.NewHttpServer(opt.GrpcAddr, services)

	defer func() {
		grpcServer.Stop()
		httpServer.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := grpcServer.Serve(opt.GrpcAddr)
		if err != nil {
			errChan <- fmt.Errorf("grpc server error: %w", err)
		}

		log.Info("grpc server stopped")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := httpServer.Serve(opt.HttpAddr)
		if err != nil {
			errChan <- fmt.Errorf("http server error: %w", err)
		}

		log.Info("http server stopped")
	}()

	<-ctx.Done()
}
