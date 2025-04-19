package otlp

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	promBridge "go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Metric interface {
	Shutdown(ctx context.Context) error
}

type metricImpl struct {
	e metric.Exporter
	r metric.Reader
}

func NewMetric(opts ...OptionFunc) (Metric, error) {
	m := &metricImpl{}
	opt := NewOption(opts...)

	return m, m.setupMetric(opt)
}

func (m *metricImpl) setupMetric(opt Option) (err error) {
	m.e, err = newMetricExporter(opt)
	if err != nil {
		return err
	}

	if opt.Gather != nil {
		m.r = NewPrometheusBridgeReader(m.e, opt.MetricInterval, opt.Gather)
	} else {
		m.r = NewPeriodicReader(m.e, opt.MetricInterval)
	}

	setupMeterProvider(m.r, opt.Resource)
	return nil
}

func (m *metricImpl) Shutdown(ctx context.Context) (err error) {
	if m.e != nil {
		if err = m.e.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown metric exporter: %w", err)
		}
	}

	if m.r != nil {
		if err = m.r.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown metric reader: %w", err)
		}
	}

	return nil
}

func setupMeterProvider(r metric.Reader, res *resource.Resource) {
	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(r),
	)

	otel.SetMeterProvider(provider)
}

func newMetricExporter(opt Option) (metric.Exporter, error) {
	switch {
	case opt.GrpcEndpoint != "":
		return otlpmetricgrpc.New(context.Background(), metricGrpcOption(opt)...)
	case opt.HTTPEndpoint != "":
		return otlpmetrichttp.New(context.Background(), metricHttpOption(opt)...)
	default:
		return nil, nil
	}
}

// NewPeriodicReader creates a new periodic reader with the given exporter and options.
func NewPeriodicReader(exp metric.Exporter, interval time.Duration) metric.Reader {
	return metric.NewPeriodicReader(exp, metric.WithInterval(interval))
}

// NewPrometheusBridgeReader creates a new periodic reader with prometheus bridge producer.
func NewPrometheusBridgeReader(exp metric.Exporter, interval time.Duration, gatherer prometheus.Gatherer) metric.Reader {
	producer := promBridge.NewMetricProducer(promBridge.WithGatherer(gatherer))
	return metric.NewPeriodicReader(exp, metric.WithInterval(interval), metric.WithProducer(producer))
}

func metricGrpcOption(opt Option) (opts []otlpmetricgrpc.Option) {
	opts = append(opts, otlpmetricgrpc.WithEndpointURL(opt.GrpcEndpoint))

	if opt.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if opt.Gzip {
		opts = append(opts, otlpmetricgrpc.WithCompressor("gzip"))
	}
	if opt.Timeout > 0 {
		opts = append(opts, otlpmetricgrpc.WithTimeout(opt.Timeout))
	}
	if opt.Header != nil {
		opts = append(opts, otlpmetricgrpc.WithHeaders(opt.Header))
	}

	return
}

func metricHttpOption(opt Option) (opts []otlpmetrichttp.Option) {
	opts = append(opts, otlpmetrichttp.WithEndpointURL(opt.HTTPEndpoint))

	if opt.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if opt.Gzip {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}
	if opt.Timeout > 0 {
		opts = append(opts, otlpmetrichttp.WithTimeout(opt.Timeout))
	}
	if opt.Header != nil {
		opts = append(opts, otlpmetrichttp.WithHeaders(opt.Header))
	}

	return
}
