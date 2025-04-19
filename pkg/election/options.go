package election

const (
	DefaultPrefix string = "election"
	DefaultTTL    int    = 10
)

type Options struct {
	// Prefix is the prefix for the election keys. Default is "election".
	Prefix string
	// TTL is the time-to-live for the election keys. Default is 10 seconds.
	TTL int
	// LeaderVal is the value to write for the leader key. Default is the json of node.NodeInfo.
	LeaderVal string
}

type Option func(*Options)

func ParseOptions(opts ...Option) *Options {
	// Default options
	o := &Options{
		Prefix: DefaultPrefix,
		TTL:    DefaultTTL,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

func WithTTL(ttl int) Option {
	return func(o *Options) {
		o.TTL = ttl
	}
}

func WithLeaderVal(val string) Option {
	return func(o *Options) {
		o.LeaderVal = val
	}
}
