package global

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	commconfig "xdp-banner/pkg/config"
	errors "xdp-banner/pkg/errors"
	random_string "xdp-banner/pkg/random"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/* Controller Config Struct */

type EtcdAuthentication struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type EtcdOptions struct {
	Endpoints      []string           `mapstructure:"endpoints"`
	Authentication EtcdAuthentication `mapstructure:"authentication"`
	DialTimeout    time.Duration      `mapstructure:"dialTimeout"`
	RequestTimeout time.Duration      `mapstructure:"requestTimeout"`
	LeaseTTL       time.Duration      `mapstructure:"leaseTTL"`
	ElectionKey    string             `mapstructure:"electionKey"`
}

func (e *EtcdOptions) Check() error {

	// etcd spec validate
	// endpoint 不需要验证，直接用来连接即可，这里只需要判空

	if len(e.Endpoints) == 0 {
		return fmt.Errorf("etcd endpoints is required")
	}

	//	if e.Endpoints == nil {
	//		return errors.NewInputError("Etcd endpoints is not specified.Check your config", nil)
	//	}

	if e.Authentication.Enabled &&
		(e.Authentication.Username == "" ||
			e.Authentication.Password == "") {
		return errors.NewInputError("Etcd authentication is enabled,but associating info is not specified.Check your config")
	}

	// etcd dialtimeout 验证，判空
	if e.DialTimeout == 0 {
		return errors.NewInputError("Etcd dialtimeout is not specified.Check your config")
	}

	// etcd requesttimeout 验证，判空
	if e.RequestTimeout == 0 {
		return errors.NewInputError("Etcd requesttimeout is not specified.Check your config")
	}

	// etcd leasettl 验证，判空
	if e.LeaseTTL == 0 {
		return errors.NewInputError("Etcd leasettl is not specified.Check your config")
	}

	// etcd electionkey 验证，判空
	if e.ElectionKey == "" {
		return errors.NewInputError("Etcd electionkey is not specified.Check your config")
	}

	return nil
}

func (e *EtcdOptions) SetFlags(cmd *cobra.Command) {
	cmdPrefix := "etcd-"
	cmd.Flags().StringSliceVar(&e.Endpoints, cmdPrefix+"endpoints", e.Endpoints, "etcd endpoints")
	cmd.Flags().BoolVar(&e.Authentication.Enabled, cmdPrefix+"auth-enabled", false, "etcd auth enabled")
	cmd.Flags().StringVar(&e.Authentication.Username, cmdPrefix+"username", e.Authentication.Username, "etcd username")
	cmd.Flags().StringVar(&e.Authentication.Password, cmdPrefix+"password", e.Authentication.Password, "etcd password")
	cmd.Flags().DurationVar(&e.DialTimeout, cmdPrefix+"dialTimeout", e.DialTimeout, "etcd dialtimeout")
	cmd.Flags().DurationVar(&e.RequestTimeout, cmdPrefix+"requestTimeout", e.RequestTimeout, "etcd password")
	cmd.Flags().DurationVar(&e.LeaseTTL, cmdPrefix+"leaseTTL", e.LeaseTTL, "etcd password")
	cmd.Flags().StringVar(&e.ElectionKey, cmdPrefix+"electionKey", e.ElectionKey, "etcd password")
}

type MetricOptions struct {
	Enabled        bool   `mapstructure:"enabled"`
	PrometheusPort uint32 `mapstructure:"port"`
}

func (m *MetricOptions) Check() error {

	// metric spec validate
	// metric prometheusPort 验证，判空
	if m.Enabled && m.PrometheusPort == 0 {
		return errors.NewInputError("Prometheus port is not specified.Check your config")
	}
	return nil
}

func (m *MetricOptions) SetFlags(cmd *cobra.Command) {
	cmdPrefix := "metric-"
	cmd.Flags().BoolVar(&m.Enabled, cmdPrefix+"enabled", false, "metirc enabled")
	cmd.Flags().Uint32Var(&m.PrometheusPort, cmdPrefix+"prometheusPort", m.PrometheusPort, "the port metric exports")
}

type OtelConfig struct {
	Endpoint    string  `mapstructure:"endpoint"`
	ServiceName string  `mapstructure:"serviceName"`
	SampleRatio float64 `mapstructure:"sampleRatio"`
	Exporter    string  `mapstructure:"exporter"`
}

type TraceOptions struct {
	Enabled bool       `mapstructure:"enabled"`
	Otel    OtelConfig `mapstructure:"otel"`
}

const (
	ExporterJaeger = "jaeger"
	ExporterOTLP   = "otlp"
	ExporterZipkin = "zipkin"
)

func (t *TraceOptions) Check() error {

	// trace spec validate
	// otel endpoint 验证，判空
	if t.Enabled && t.Otel.Endpoint == "" {
		return errors.NewInputError("Otel endpoints is not specified.Check your config")
	}
	// otel serviceName 验证，判空
	if t.Enabled && t.Otel.ServiceName == "" {
		return errors.NewInputError("Otel serviceName is not specified.Check your config")
	}
	// otel sampleRatio 验证，判空
	if t.Enabled && t.Otel.SampleRatio == 0 {
		return errors.NewInputError("Otel sampleRatio is not specified.Check your config")
	}
	// otel sampleRatio 范围检验
	if t.Enabled && (t.Otel.SampleRatio < 0 || t.Otel.SampleRatio > 1) {
		return errors.NewInputError("Otel sampleRatio is in a wrong value.Check your config")
	}
	// otel exporter 验证，判空
	switch t.Otel.Exporter {
	case ExporterJaeger, ExporterOTLP, ExporterZipkin:
	default:
		return errors.NewInputError("Otel exporter is invalid your config")
	}

	return nil

}

func (t *TraceOptions) SetFlags(cmd *cobra.Command) {
	cmdPrefix := "trace-"
	cmd.Flags().BoolVar(&t.Enabled, cmdPrefix+"enabled", false, "etcd auth enabled")
	cmd.Flags().StringVar(&t.Otel.Endpoint, cmdPrefix+"otel-endpoint", t.Otel.Endpoint, "etcd password")
	cmd.Flags().Float64Var(&t.Otel.SampleRatio, cmdPrefix+"dial-timeout", t.Otel.SampleRatio, "otel sample ratio")
	cmd.Flags().StringVar(&t.Otel.Exporter, cmdPrefix+"exporter", t.Otel.Exporter, "otel exporter")
	cmd.Flags().StringVar(&t.Otel.ServiceName, cmdPrefix+"service-Nname", t.Otel.ServiceName, "etcd password")
}

type LogOptions struct {
	Enabled bool   `mapstructure:"enabled"`
	Level   string `mapstructure:"level"`
	Path    string `mapstructure:"path"`
}

const (
	Info  string = "info"
	Warn  string = "warn"
	Debug string = "debug"
)

func (l *LogOptions) Check() error {

	// log spec validate
	// log level 验证，判空
	if l.Enabled && l.Level == "" {
		return errors.NewInputError("Log level is not specified.Check your config")
	}

	switch l.Level {
	case Info, Warn, Debug:
	default:
		return errors.NewInputError("Log level should in 'info', 'warn' or 'debug'.Check your config")
	}

	// log path 验证，判空
	if l.Enabled && l.Path == "" {
		return errors.NewInputError("Log path is not specified.Check your config")
	}

	return nil
}

func (l *LogOptions) SetFlags(cmd *cobra.Command) {
	cmdPrefix := "log-"
	cmd.Flags().BoolVar(&l.Enabled, cmdPrefix+"enabled", false, "log enabled")
	cmd.Flags().StringVar(&l.Level, cmdPrefix+"level", l.Level, "log level")
	cmd.Flags().StringVar(&l.Path, cmdPrefix+"path", l.Path, "log path")
}

type ControllerOptions struct {
	ControllerName string        `mapstructure:"controllerName"`
	Etcd           EtcdOptions   `mapstructure:"etcdSpec"`
	Metric         MetricOptions `mapstructure:"metric"`
	Trace          TraceOptions  `mapstructure:"trace"`
	Log            LogOptions    `mapstructure:"log"`
}

func DefaultOption() *ControllerOptions {
	return &ControllerOptions{
		ControllerName: "xdp-banner-" + random_string.RandomString(5),
		Etcd: EtcdOptions{
			Endpoints: []string{"http://127.0.0.1:2379", "http://127.0.0.1:2380"},
			Authentication: EtcdAuthentication{
				Enabled:  false,
				Username: "admin",
				Password: "password123",
			},
			DialTimeout:    5 * time.Second,
			RequestTimeout: 2 * time.Second,
			LeaseTTL:       10 * time.Second,
			ElectionKey:    "/election/key",
		},
		Metric: MetricOptions{
			Enabled:        false,
			PrometheusPort: 9090,
		},
		Trace: TraceOptions{
			Enabled: false,
			Otel: OtelConfig{
				Endpoint:    "localhost:4317",
				ServiceName: "xdp-banner",
				SampleRatio: 1.0,
				Exporter:    "otlp",
			},
		},
		Log: LogOptions{
			Enabled: true,
			Level:   Info,
			Path:    "/var/log/xdp-banner.log",
		},
	}
}

func (e *ControllerOptions) Check() error {

	err := e.Etcd.Check()
	if err != nil {
		return err
	}

	err = e.Metric.Check()
	if err != nil {
		return err
	}

	err = e.Trace.Check()
	if err != nil {
		return err
	}
	err = e.Log.Check()

	if err != nil {
		return err
	}
	return nil
}

func (e *ControllerOptions) SetFlags(cmd *cobra.Command) {
	e.Etcd.SetFlags(cmd)
	e.Metric.SetFlags(cmd)
	e.Trace.SetFlags(cmd)
	e.Log.SetFlags(cmd)
}

func LoadConfig(config_path string) (*viper.Viper, *ControllerOptions, error) {

	var controllerConfig ControllerOptions

	// 默认 file 的后缀为 .yaml
	dir, file := filepath.Split(config_path)
	filename := strings.Split(file, ".")
	viperSample, err := commconfig.NewFromConfig(filename[0], dir, &controllerConfig)

	if err != nil {
		return nil, nil, err
	}

	err = controllerConfig.Check()
	if err != nil {
		return nil, nil, err
	}

	return viperSample, &controllerConfig, err
}
