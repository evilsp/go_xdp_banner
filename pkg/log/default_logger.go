package log

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// newDefaultProductionLog configures a custom log that is
// intended for use by default if no other log is specified
// in a config. It writes to stderr, uses the console encoder,
// and enables INFO-level logs and higher.
func newDefaultProductionLog() zap.Logger {
	bl := new(BaseLog)
	bl.writerOpener = StderrWriter{}
	bl.writer, _ = bl.writerOpener.OpenWriter()

	bl.encoder = newDefaultProductionLogEncoder(bl.writerOpener)
	bl.levelEnabler = zapcore.DebugLevel

	bl.buildCore()

	logger := zap.New(bl.core)

	// capture logs from other libraries which
	// may not be using zap logging directly
	_ = zap.RedirectStdLog(logger)

	return *logger
}

// LogSampling configures log entry sampling.
type LogSampling struct {
	// The window over which to conduct sampling.
	Interval time.Duration `json:"interval,omitempty"`

	// Log this many entries within a given level and
	// message for each interval.
	First int `json:"first,omitempty"`

	// If more entries with the same level and message
	// are seen during the same interval, keep one in
	// this many entries until the end of the interval.
	Thereafter int `json:"thereafter,omitempty"`
}

// BaseLog contains the common logging parameters for logging.
type BaseLog struct {
	// Level is the minimum level to emit, and is inclusive.
	// Possible levels: DEBUG, INFO, WARN, ERROR, PANIC, and FATAL
	Level string `json:"level,omitempty"`

	// Sampling configures log entry sampling. If enabled,
	// only some log entries will be emitted. This is useful
	// for improving performance on extremely high-pressure
	// servers.
	Sampling *LogSampling `json:"sampling,omitempty"`

	// If true, the log entry will include the caller's
	// file name and line number. Default off.
	WithCaller bool `json:"with_caller,omitempty"`

	// If non-zero, and `with_caller` is true, this many
	// stack frames will be skipped when determining the
	// caller. Default 0.
	WithCallerSkip int `json:"with_caller_skip,omitempty"`

	// If not empty, the log entry will include a stack trace
	// for all logs at the given level or higher. See `level`
	// for possible values. Default off.
	WithStacktrace string `json:"with_stacktrace,omitempty"`

	writerOpener WriterOpener
	writer       io.WriteCloser
	encoder      zapcore.Encoder
	levelEnabler zapcore.LevelEnabler
	core         zapcore.Core
}

func (cl *BaseLog) buildCore() {
	// logs which only discard their output don't need
	// to perform encoding or any other processing steps
	// at all, so just shortcut to a nop core instead
	if _, ok := cl.writerOpener.(*DiscardWriter); ok {
		cl.core = zapcore.NewNopCore()
		return
	}
	c := zapcore.NewCore(
		cl.encoder,
		zapcore.AddSync(cl.writer),
		cl.levelEnabler,
	)
	if cl.Sampling != nil {
		if cl.Sampling.Interval == 0 {
			cl.Sampling.Interval = 1 * time.Second
		}
		if cl.Sampling.First == 0 {
			cl.Sampling.First = 100
		}
		if cl.Sampling.Thereafter == 0 {
			cl.Sampling.Thereafter = 100
		}
		c = zapcore.NewSamplerWithOptions(c, cl.Sampling.Interval,
			cl.Sampling.First, cl.Sampling.Thereafter)
	}
	cl.core = c
}

func newDefaultProductionLogEncoder(wo WriterOpener) zapcore.Encoder {
	encCfg := zap.NewProductionEncoderConfig()
	if IsWriterStandardStream(wo) && term.IsTerminal(int(os.Stderr.Fd())) {
		// if interactive terminal, make output more human-readable by default
		encCfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format("2006/01/02 15:04:05.000"))
		}

		coloringEnabled := os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "xterm-mono"
		if coloringEnabled {
			encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		return zapcore.NewConsoleEncoder(encCfg)
	}
	return zapcore.NewJSONEncoder(encCfg)
}
