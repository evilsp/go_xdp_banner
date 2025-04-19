package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"xdp-banner/orch/cmd/global"
	"xdp-banner/orch/logic"
	"xdp-banner/orch/storage"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/otlp"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.uber.org/zap"
)

func Cmd(parentOpt *global.ControllerOptions) *cobra.Command {
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
	setupOtlp(opt)
	setupOtelInstrumentation()

	// Cli init
	err := global.CreateGlobalEtcdInstance(opt.Parent)

	if err != nil {
		log.Fatal("failed to connect to etcd", zap.Error(err))
	}

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := storage.New(ctx, global.Cli)
	logic := logic.New(storage)

	go runController(ctx, global.Cli, logic, wg, errChan)
	go runServer(ctx, opt, logic, wg, errChan)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer func() {
		signal.Stop(signals)
	}()

	select {
	case err := <-errChan:
		log.Error("server error, existing", log.ErrorField(err))
	case sig := <-signals:
		log.Info("receive exit signal, exiting", log.AnyField("signal", sig))
	}

	cancel()
	wg.Wait()
}

func setupOtlp(opt *Option) {
	setupOtlpLog(opt)
	setupOtlpMetric(opt)
}

func setupOtlpLog(opt *Option) {
	res, err := otlp.DefaultResource("orch")
	if err != nil {
		log.Fatal("otlp resource error", zap.Error(err))
	}

	core, err := otlp.NewLog(
		otlp.WithName("orch"),
		otlp.WithHTTPEndpoint(opt.Otlp.LoggerHTTPEndpoint),
		otlp.WithGrpcEndpoint(opt.Otlp.LoggerGrpcEndpoint),
		otlp.WithResource(res),
		otlp.WithInsecure(opt.Otlp.Insecure),
		otlp.WithGzip(opt.Otlp.Gzip),
		otlp.WithHeader(opt.Otlp.Header),
		otlp.WithTimeout(opt.Otlp.Timeout),
	)
	if err != nil {
		log.Fatal("otlp log error", zap.Error(err))
	}

	core = log.NewTreeWithDefaultLogger(core)
	log.SetGlobalLogger(zap.New(core))
}

func setupOtlpMetric(opt *Option) {
	res, err := otlp.DefaultResource("orch")
	if err != nil {
		log.Fatal("otlp resource error", zap.Error(err))
	}

	_, err = otlp.NewMetric(
		otlp.WithHTTPEndpoint(opt.Otlp.MetricHTTPEndpoint),
		otlp.WithGrpcEndpoint(opt.Otlp.MetricGrpcEndpoint),
		otlp.WithResource(res),
		otlp.WithInsecure(opt.Otlp.Insecure),
		otlp.WithGzip(opt.Otlp.Gzip),
		otlp.WithHeader(opt.Otlp.Header),
		otlp.WithTimeout(opt.Otlp.Timeout),
	)
	if err != nil {
		log.Fatal("otlp metric error", zap.Error(err))
	}
}

func setupOtelInstrumentation() {
	err := runtime.Start()
	if err != nil {
		log.Warn("failed to start runtime instrumentation, we will lack runtime metrics", zap.Error(err))
	}
}
