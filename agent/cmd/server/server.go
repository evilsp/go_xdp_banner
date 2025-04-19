package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"xdp-banner/agent/cmd/global"
	"xdp-banner/agent/internal/client"
	"xdp-banner/agent/internal/icert"
	"xdp-banner/agent/internal/statusfsm"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/node"
	"xdp-banner/pkg/otlp"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Cmd(parentOpt *global.Option) *cobra.Command {
	opt := DefaultOption(parentOpt)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opt.Check(); err != nil {
				return err
			}

			run(opt)
			return nil
		},
	}

	opt.SetFlags(cmd)

	return cmd
}

func run(opt *Option) {
	setupOtlpLog(opt)

	cred, err := NewCredits()
	if err != nil {
		log.Fatal("create credentials", zap.Error(err))
	}

	cli, err := client.New(opt.Parent.Orch.Endpoints,
		grpc.WithTransportCredentials(cred),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // 启用 Round Robin
	)
	if err != nil {
		log.Fatal("create client", zap.Error(err))
	}

	client.SetupReporter(cli, opt.ReportInterval)
	GatherBasicInfo(opt)
	client.StartReporter()

	controller := initControllerCtx(cli)
	if err != nil {
		log.Fatal("create credentials", zap.Error(err))
	}

	fsm := statusfsm.New(
		ErrorWrapper(controller.Start),
		ErrorWrapper(controller.Stop),
		ErrorWrapper(controller.Reload),
	)

	grpcServices := NewGrpcServices(fsm)
	grpcServer := NewGrpcServer(grpcServices, cred)

	if err := grpcServer.Serve(opt.GrpcAddr); err != nil {
		log.Fatal("serve", log.ErrorField(err))
	}
}

// GatherBasicInfo gathers basic info about this agent, and set to reporter
func GatherBasicInfo(opt *Option) {
	// Node name
	name, err := node.Name()
	if err != nil {
		log.FatalE("get node name", err)
	}
	client.SetName(name)

	// grpc address
	ip, err := node.DefaultIP()
	if err != nil {
		log.FatalE("get default ip", err)
	}

	_, port, err := net.SplitHostPort(opt.GrpcAddr)
	if err != nil {
		log.FatalE("extract port from grpc listen address", err)
	}
	client.SetGrpcEndpoint(fmt.Sprintf("%s:%s", ip, port))
}

func NewCredits() (credentials.TransportCredentials, error) {
	// load ca
	caCert, err := icert.GetCA()
	if err != nil {
		return nil, fmt.Errorf("read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("add CA certificate")
	}

	// load cert
	cert, err := icert.GetCertPair()
	if err != nil {
		return nil, fmt.Errorf("load node cert pair: %w", err)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}), nil
}

func setupOtlpLog(opt *Option) {
	res, err := otlp.DefaultResource("agent")
	if err != nil {
		log.Fatal("otlp resource error", zap.Error(err))
	}

	core, err := otlp.NewLog(
		otlp.WithName("agent"),
		otlp.WithHTTPEndpoint(opt.Otlp.LoggerHTTPEndpoint),
		otlp.WithGrpcEndpoint(opt.Otlp.LoggerGrpcEndpoint),
		otlp.WithResource(res),
		otlp.WithInsecure(opt.Otlp.Insecure),
		otlp.WithGzip(opt.Otlp.Insecure),
		otlp.WithHeader(opt.Otlp.Header),
		otlp.WithTimeout(opt.Otlp.Timeout),
	)
	if err != nil {
		log.Fatal("otlp log error", zap.Error(err))
	}

	core = log.NewTreeWithDefaultLogger(core)
	log.SetGlobalLogger(zap.New(core))
}
