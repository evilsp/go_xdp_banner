package otlp

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Option holds configuration parameters.
type Option struct {
	// Gzip indicates whether to enable gzip compression. Default: false
	Gzip bool
	// Insecure indicates whether to skip TLS. Default: false
	Insecure bool
	// Timeout sets the maximum duration for an export attempt. After reaching this limit, export is abandoned.
	Timeout time.Duration
	// MetricInterval is the interval for scraping metrics. Default: 15 seconds
	MetricInterval time.Duration
	// Header stores custom headers to be sent to the OTLP server.
	Header map[string]string
	// GrpcEndpoint is the gRPC endpoint for OTLP. If left empty, gRPC is not used.
	GrpcEndpoint string
	// HTTPEndpoint is the HTTP endpoint for OTLP. If left empty, HTTP is not used.
	HTTPEndpoint string
	// Resource is the resource attributes for the OTLP exporter.
	Resource *resource.Resource
	// Gatherer use prometheus gather as the additional source of metrics, only used for metric.
	// so a prometheus insrumented app can use this to export metrics to OTLP.
	Gather prometheus.Gatherer
	//Name is the name of the logger
	Name string
}

// OptionFunc defines a function that configures an Option.
type OptionFunc func(*Option)

// NewOption creates an Option with the provided OptionFunc values.
func NewOption(opts ...OptionFunc) Option {
	o := &Option{
		Gzip:           false,
		Insecure:       false,
		Timeout:        10 * time.Second,
		MetricInterval: 15 * time.Second,
		Header:         make(map[string]string),
		GrpcEndpoint:   "",
		HTTPEndpoint:   "",
	}
	for _, fn := range opts {
		fn(o)
	}
	return *o
}

// WithGzip sets whether to enable gzip compression.
func WithGzip(enable bool) OptionFunc {
	return func(o *Option) {
		o.Gzip = enable
	}
}

// WithInsecure sets whether to skip TLS.
func WithInsecure(enable bool) OptionFunc {
	return func(o *Option) {
		o.Insecure = enable
	}
}

// WithTimeout sets the maximum duration for an export attempt.
func WithTimeout(d time.Duration) OptionFunc {
	return func(o *Option) {
		o.Timeout = d
	}
}

// WithMetricInterval sets the interval for scraping metrics.
func WithMetricInterval(d time.Duration) OptionFunc {
	return func(o *Option) {
		o.MetricInterval = d
	}
}

// WithHeader sets custom headers for the OTLP server.
func WithHeader(h map[string]string) OptionFunc {
	return func(o *Option) {
		o.Header = h
	}
}

// WithGrpcEndpoint sets the gRPC endpoint for OTLP.
func WithGrpcEndpoint(endpoint string) OptionFunc {
	return func(o *Option) {
		o.GrpcEndpoint = endpoint
	}
}

// WithHTTPEndpoint sets the HTTP endpoint for OTLP.
func WithHTTPEndpoint(endpoint string) OptionFunc {
	return func(o *Option) {
		o.HTTPEndpoint = endpoint
	}
}

// WithGatherer sets the prometheus gatherer for OTLP.
func WithGatherer(g prometheus.Gatherer) OptionFunc {
	return func(o *Option) {
		o.Gather = g
	}
}

// WithResource sets the resource attributes for the OTLP exporter.
func WithResource(r *resource.Resource) OptionFunc {
	return func(o *Option) {
		o.Resource = r
	}
}

// WithName sets the name of the logger.
func WithName(name string) OptionFunc {
	return func(o *Option) {
		o.Name = name
	}
}
