package option

import (
	"time"

	"github.com/spf13/cobra"
)

type OtlpOption struct {
	Insecure           bool
	Gzip               bool
	Timeout            time.Duration
	MetricInterval     time.Duration
	Header             map[string]string
	MetricGrpcEndpoint string
	MetricHTTPEndpoint string
	LoggerGrpcEndpoint string
	LoggerHTTPEndpoint string
}

func (o *OtlpOption) SetFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Insecure, "otel.insecure", o.Insecure, "enable insecure connection")
	cmd.Flags().BoolVar(&o.Gzip, "otel.gzip", o.Gzip, "enable gzip compression")
	cmd.Flags().DurationVar(&o.Timeout, "otel.timeout", o.Timeout, "set timeout for the connection")
	cmd.Flags().DurationVar(&o.MetricInterval, "otel.metric-interval", o.MetricInterval, "set interval for the metrics")
	cmd.Flags().StringToStringVar(&o.Header, "otel.header", o.Header, "set header for the connection")
	cmd.Flags().StringVar(&o.MetricGrpcEndpoint, "otel.metric-grpc-endpoint", o.MetricGrpcEndpoint, "set grpc endpoint for the metrics")
	cmd.Flags().StringVar(&o.MetricHTTPEndpoint, "otel.metric-http-endpoint", o.MetricHTTPEndpoint, "set http endpoint for the metrics")
	cmd.Flags().StringVar(&o.LoggerGrpcEndpoint, "otel.logger-grpc-endpoint", o.LoggerGrpcEndpoint, "set grpc endpoint for the logger")
	cmd.Flags().StringVar(&o.LoggerHTTPEndpoint, "otel.logger-http-endpoint", o.LoggerHTTPEndpoint, "set http endpoint for the logger")
}

func DefaultOtelOption() *OtlpOption {
	return &OtlpOption{
		Insecure:           true,
		Gzip:               false,
		Timeout:            5 * time.Second,
		MetricInterval:     15 * time.Second,
		Header:             map[string]string{},
		MetricGrpcEndpoint: "",
		MetricHTTPEndpoint: "http://prometheus-server.xdp-banner.svc.cluster.local/api/v1/otlp/v1/metrics",
		LoggerGrpcEndpoint: "",
		LoggerHTTPEndpoint: "http://loki-gateway.xdp-banner.svc.cluster.local/otlp/v1/logs",
	}
}
