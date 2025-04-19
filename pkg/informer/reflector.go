package informer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"

	"xdp-banner/pkg/utils/clock"
	"xdp-banner/pkg/wait"
)

var (
	// We try to spread the load on apiserver by setting timeouts for
	// watch requests - it is random in [minWatchTimeout, 2*minWatchTimeout].
	defaultMinWatchTimeout = 5 * time.Minute
)

// Reflector watches a specified resource and causes all changes to be reflected in the given store.
type Reflector struct {
	// name identifies this reflector. By default, it will be a file:line if possible.
	name string
	// The prefix is the resource being watched.
	prefix string
	// The destination to sync up with the watch source
	store Store
	// listerWatcher is used to perform lists and watches.
	listerWatcher etcd.ListerWatcher
	// backoff manages backoff of ListWatch
	backoffManager wait.BackoffManager
	// minWatchTimeout defines the minimum timeout for watch requests.
	minWatchTimeout time.Duration
	// clock allows tests to manipulate time
	clock clock.Clock
	// lastSyncResourceVersion is the resource version token last
	// observed when doing a sync with the underlying store
	// it is thread safe, but not synchronized with the underlying store
	lastSyncResourceVersion int64
	// isLastSyncResourceVersionUnavailable is true if the previous list or watch request with
	// lastSyncResourceVersion failed with an "expired" or "too large resource version" error.
	isLastSyncResourceVersionUnavailable bool
	// lastSyncResourceVersionMutex guards read/write access to lastSyncResourceVersion
	lastSyncResourceVersionMutex sync.RWMutex
	// Called whenever the ListAndWatch drops the connection with an error.
	watchErrorHandler WatchErrorHandler
	// WatchListPageSize is the requested chunk size of initial and resync watch lists.
	// If unset, for consistent reads (RV="") or reads that opt-into arbitrarily old data
	// (RV="0") it will default to pager.PageSize, for the rest (RV != "" && RV != "0")
	// it will turn off pagination to allow serving them from watch cache.
	// NOTE: It should be used carefully as paginated lists are always served directly from
	// etcd, which is significantly less efficient and may lead to serious performance and
	// scalability problems.
	WatchListPageSize int64
	// MaxInternalErrorRetryDuration defines how long we should retry internal errors returned by watch.
	MaxInternalErrorRetryDuration time.Duration
}

func (r *Reflector) Name() string {
	return r.name
}

func (r *Reflector) TypeDescription() string {
	return r.prefix
}

// ResourceVersionUpdater is an interface that allows store implementation to
// track the current resource version of the reflector. This is especially
// important if storage bookmarks are enabled.
type ResourceVersionUpdater interface {
	// UpdateResourceVersion is called each time current resource version of the reflector
	// is updated.
	UpdateResourceVersion(resourceVersion int64)
}

// The WatchErrorHandler is called whenever ListAndWatch drops the
// connection with an error. After calling this handler, the informer
// will backoff and retry.
//
// The default implementation looks at the error type and tries to log
// the error message at an appropriate level.
//
// Implementations of this handler may display the error message in other
// ways. Implementations should return quickly - any expensive processing
// should be offloaded.
type WatchErrorHandler func(r *Reflector, err error)

// DefaultWatchErrorHandler is the default implementation of WatchErrorHandler
func DefaultWatchErrorHandler(r *Reflector, err error) {
	switch {
	case err == io.EOF:
		// watch closed normally
	case err == io.ErrUnexpectedEOF:
		log.Info("Watch closed with unexpected EOF", log.StringField("name", r.name), log.StringField("prefix", r.prefix), log.ErrorField(err))
	default:
		log.Info("Failed to watch", log.StringField("name", r.name), log.StringField("prefix", r.prefix), log.ErrorField(err))
	}
}

// NewReflector creates a new Reflector with its name defaulted to the closest source_file.go:line in the call stack
// that is outside this package. See NewReflectorWithOptions for further information.
func NewReflector(lw etcd.ListerWatcher, name string, prefix string, store Store) *Reflector {
	clock := clock.RealClock{}

	return &Reflector{
		name:            name,
		minWatchTimeout: defaultMinWatchTimeout,
		prefix:          prefix,
		listerWatcher:   lw,
		store:           store,
		// We used to make the call every 1sec (1 QPS), the goal here is to achieve ~98% traffic reduction when
		// API server is not healthy. With these parameters, backoff will stop at [30,60) sec interval which is
		// 0.22 QPS. If we don't backoff for 2min, assume API server is healthy and we reset the backoff.
		backoffManager:    wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, clock),
		clock:             clock,
		watchErrorHandler: WatchErrorHandler(DefaultWatchErrorHandler),
	}

}

// Run repeatedly uses the reflector's ListAndWatch to fetch all the
// objects and subsequent deltas.
// Run will exit when stopCh is closed.
func (r *Reflector) Run(ctx context.Context) {
	log.Info("Starting reflector", log.StringField("name", r.name), log.StringField("prefix", r.prefix))
	wait.BackoffUntil(func() {
		if err := r.ListAndWatch(ctx); err != nil {
			r.watchErrorHandler(r, err)
		}
	}, r.backoffManager, true, ctx)
	log.Info("reflector stopped", log.StringField("name", r.name))
}

var (
	// Used to indicate that watching stopped because of a signal from the stop
	// channel passed in from a client of the reflector.
	errorStopRequested = errors.New("stop requested")
)

// ListAndWatch first lists all items and get the resource version at the moment of call,
// and then use the resource version to watch.
// It returns error if ListAndWatch didn't even try to initialize watch.
func (r *Reflector) ListAndWatch(ctx context.Context) error {
	log.Info("Listing and watching", log.StringField("name", r.name), log.StringField("prefix", r.prefix))
	var err error
	var w etcd.WatchController

	err = r.list(ctx)
	if err != nil {
		return err
	}

	log.Info("Caches populated", log.StringField("name", r.name), log.StringField("prefix", r.prefix))
	return r.watch(ctx, w)
}

// watch simply starts a watch request with the server.
func (r *Reflector) watch(ctx context.Context, w etcd.WatchController) error {
	var err error

	for {
		// give the stopCh a chance to stop the loop, even in case of continue statements further down on errors
		select {
		case <-ctx.Done():
			// we can only end up here when the stopCh
			// was closed after a successful watchlist or list request
			if w != nil {
				w.Stop()
			}
			return nil
		default:
		}

		// start the clock before sending the request, since some proxies won't flush headers until after the first watch event is sent
		start := r.clock.Now()

		if w == nil {
			options := etcd.WatchOption{
				Prefix:   r.prefix,
				Revision: r.rewatchResourceVersion(),
			}
			w, err = r.listerWatcher.Watch(options)
			if err != nil {
				return err
			}
		}

		err = handleWatch(ctx, start, w, r.store, r.name, r.prefix, r.setLastSyncResourceVersion, r.clock)
		// Ensure that watch will not be reused across iterations.
		w.Stop()
		w = nil
		if err != nil {
			if !errors.Is(err, errorStopRequested) {
				log.Error("Error while watching", log.StringField("name", r.name), log.StringField("prefix", r.prefix), log.ErrorField(err))
				continue
			}
			return nil
		}
	}
}

// list simply lists all items and records a resource version obtained from the server at the moment of the call.
// the resource version can be used for further progress notification (aka. watch).
func (r *Reflector) list(ctx context.Context) error {
	var resourceVersion int64
	options := etcd.ListOption{
		Size:     r.WatchListPageSize,
		Prefix:   r.prefix,
		Revision: r.relistResourceVersion(),
	}

	var list etcd.PagedList
	var err error
	listCh := make(chan struct{}, 1)
	panicCh := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		pager := etcd.NewListPager(r.listerWatcher.List)
		list, err = pager.List(ctx, options)
		close(listCh)
	}()
	select {
	case <-ctx.Done():
		return nil
	case r := <-panicCh:
		panic(r)
	case <-listCh:
	}
	if err != nil {
		log.Warn("failed to list", log.StringField("name", r.name), log.StringField("prefix", r.prefix), log.ErrorField(err))
		return fmt.Errorf("failed to list %v: %w", r.prefix, err)
	}

	r.setIsLastSyncResourceVersionUnavailable(false) // list was successful
	resourceVersion = list.Revision
	if err := r.syncWith(list.Items, resourceVersion); err != nil {
		return fmt.Errorf("unable to sync list result: %v", err)
	}
	r.setLastSyncResourceVersion(resourceVersion)
	return nil
}

func (r *Reflector) syncWith(list *etcd.Items, resourceVersion int64) error {
	listIter := func(iter func(string, any) bool) {
		for k, v := range list.Iterator() {
			if !iter(k, v) {
				return
			}
		}
	}

	if err := r.store.Replace(listIter, strconv.Itoa(int(resourceVersion))); err != nil {
		return err
	}

	r.store.SyncDone()
	return nil
}

// watchHandler watches w and sets setLastSyncResourceVersion
func handleWatch(
	ctx context.Context,
	start time.Time,
	w etcd.WatchController,
	store Store,
	name string,
	prefix string,
	setLastSyncResourceVersion func(int64),
	clock clock.Clock,
) error {
	eventCount := 0

	// Stopping the watcher should be idempotent and if we return from this function there's no way
	// we're coming back in with the same watch interface.
	defer w.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			if event.Type == etcd.Error {
				return event.Value.(error)
			}
			resourceVersion := event.Revision
			switch event.Type {
			case etcd.Put:
				err := store.Update(event.Key, event.Value)
				if err != nil {
					log.Error("Unable to update watch event object to store", log.StringField("name", name), log.AnyField("value", event.Value), log.ErrorField(err))
				}
			case etcd.Delete:
				// TODO: Will any consumers need access to the "last known
				// state", which is passed in event.Object? If so, may need
				// to change this.
				err := store.Delete(event.Key)
				if err != nil {
					log.Error("Unable to delete watch event object from store", log.StringField("name", name), log.AnyField("value", event.Value), log.ErrorField(err))
				}
			default:
				log.Error("Unable to understand watch event", log.StringField("name", name), log.AnyField("value", event.Value))
			}
			setLastSyncResourceVersion(resourceVersion)
			if rvu, ok := store.(ResourceVersionUpdater); ok {
				rvu.UpdateResourceVersion(resourceVersion)
			}
			eventCount++
		}
	}

	watchDuration := clock.Since(start)
	if watchDuration < 1*time.Second && eventCount == 0 {
		return fmt.Errorf("very short watch: %s: Unexpected watch close - watch lasted less than a second and no items received", name)
	}
	log.Info("Watch close", log.StringField("name", name), log.StringField("prefix", prefix), log.DurationField("duration", watchDuration), log.IntField("count", eventCount))
	return nil
}

// LastSyncResourceVersion is the resource version observed when last sync with the underlying store
// The value returned is not synchronized with access to the underlying store and is not thread-safe
func (r *Reflector) LastSyncResourceVersion() int64 {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()
	return r.lastSyncResourceVersion
}

func (r *Reflector) setLastSyncResourceVersion(v int64) {
	r.lastSyncResourceVersionMutex.Lock()
	defer r.lastSyncResourceVersionMutex.Unlock()
	r.lastSyncResourceVersion = v
}

// relistResourceVersion determines the resource version the reflector should list or relist from.
// Returns either the lastSyncResourceVersion so that this reflector will relist with a resource
// versions no older than has already been observed in relist results or watch events, or, if the last relist resulted
// in an HTTP 410 (Gone) status code, returns "" so that the relist will use the latest resource version available in
// etcd via a quorum read.
func (r *Reflector) relistResourceVersion() int64 {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()

	if r.isLastSyncResourceVersionUnavailable {
		// Since this reflector makes paginated list requests, and all paginated list requests skip the watch cache
		// if the lastSyncResourceVersion is unavailable, we set ResourceVersion="" and list again to re-establish reflector
		// to the latest available ResourceVersion, using a consistent read from etcd.
		return 0
	}

	return r.lastSyncResourceVersion
}

// rewatchResourceVersion determines the resource version the reflector should start streaming from.
func (r *Reflector) rewatchResourceVersion() int64 {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()
	if r.isLastSyncResourceVersionUnavailable {
		// initial stream should return data at the most recent resource version.
		// the returned data must be consistent i.e. as if served from etcd via a quorum read
		return 0
	}
	return r.lastSyncResourceVersion
}

// setIsLastSyncResourceVersionUnavailable sets if the last list or watch request with lastSyncResourceVersion returned
// "expired" or "too large resource version" error.
func (r *Reflector) setIsLastSyncResourceVersionUnavailable(isUnavailable bool) {
	r.lastSyncResourceVersionMutex.Lock()
	defer r.lastSyncResourceVersionMutex.Unlock()
	r.isLastSyncResourceVersionUnavailable = isUnavailable
}

var _ clock.Ticker = &noopTicker{}

type noopTicker struct{}

func (t *noopTicker) C() <-chan time.Time { return nil }

func (t *noopTicker) Stop() {}
