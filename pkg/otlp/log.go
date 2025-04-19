package otlp

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otel "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap/zapcore"
)

func NewLog(opts ...OptionFunc) (zapcore.Core, error) {
	opt := NewOption(opts...)
	return setupZapLog(opt)
}

// setupLog sets up the logger and config opentelemetry log for the application.
func setupZapLog(opt Option) (zapcore.Core, error) {
	exp, err := newLogExporter(opt)
	if err != nil {
		return nil, err
	}

	prod := newLoggerProvider(exp, opt.Resource)
	core := newZapBridgeLogger(opt.Name, prod)

	return core, nil
}

func newZapBridgeLogger(name string, provider otel.LoggerProvider) zapcore.Core {
	return otelzap.NewCore(name, otelzap.WithLoggerProvider(provider))
}

func newLoggerProvider(exporter log.Exporter, res *resource.Resource) otel.LoggerProvider {
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
		log.WithResource(res),
	)

	return loggerProvider
}

func newLogExporter(opt Option) (log.Exporter, error) {
	if opt.GrpcEndpoint != "" {
		e, err := otlploggrpc.New(context.Background(), logGrpcOption(opt)...)
		if err != nil {
			return nil, fmt.Errorf("create otlp log exporter failed: %w", err)
		}
		return e, nil
	}

	if opt.HTTPEndpoint != "" {
		e, err := otlploghttp.New(context.Background(), logHttpOption(opt)...)
		if err != nil {
			return nil, fmt.Errorf("create otlp log exporter failed: %w", err)
		}
		return e, nil
	}

	return nil, fmt.Errorf("no endpoint provided")
}

func logGrpcOption(opt Option) (opts []otlploggrpc.Option) {
	opts = append(opts, otlploggrpc.WithEndpointURL(opt.GrpcEndpoint))

	if opt.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if opt.Gzip {
		opts = append(opts, otlploggrpc.WithCompressor("gzip"))
	}
	if opt.Timeout > 0 {
		opts = append(opts, otlploggrpc.WithTimeout(opt.Timeout))
	}
	if opt.Header != nil {
		opts = append(opts, otlploggrpc.WithHeaders(opt.Header))
	}

	return
}

func logHttpOption(opt Option) (opts []otlploghttp.Option) {
	opts = append(opts, otlploghttp.WithEndpointURL(opt.HTTPEndpoint))

	if opt.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if opt.Gzip {
		opts = append(opts, otlploghttp.WithCompression(otlploghttp.GzipCompression))
	}
	if opt.Timeout > 0 {
		opts = append(opts, otlploghttp.WithTimeout(opt.Timeout))
	}
	if opt.Header != nil {
		opts = append(opts, otlploghttp.WithHeaders(opt.Header))
	}

	return
}
