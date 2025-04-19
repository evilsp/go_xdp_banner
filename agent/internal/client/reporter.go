package client

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"xdp-banner/api/orch/v1/agent/report"
	"xdp-banner/pkg/log"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var r Reporter

func SetupReporter(client Client, interval time.Duration) {
	r = NewReporter(client, interval)

	SetReporterInstance(r)
}

func StartReporter() {
	go r.Start()
}

func SetReporterInstance(reporter Reporter) {
	r = reporter
}

func SetName(name string) {
	r.SetData(NameKey, name)
}

func SetGrpcEndpoint(endpoint string) {
	r.SetData(GrpcEndpointKey, endpoint)
}

func SetConfigName(configName string) {
	r.SetData(ConfigNameKey, configName)
}

func SetPhase(phase report.Phase) {
	if phase == report.Phase_Running {
		// if the phase is running, we need to set the error to nil
		// to clear the error
		r.SetDatas([]MetricKey{Error, PhaseKey}, []any{nil, phase})
		return
	}

	r.SetData(PhaseKey, phase)
}

type ErrorTime struct {
	Message string
	RetryAt *timestamppb.Timestamp
}

func SetError(err string) {
	r.SetData(Error, ErrorTime{
		Message: err,
		RetryAt: timestamppb.New(time.Now().Add(5 * time.Second)),
	})
}

type Reporter interface {
	Start()
	Close()

	SetData(key MetricKey, value any)
	SetDatas(key []MetricKey, value []any)
}

type MetricKey int

// fieldNum is the number of fields below
const filedNum = 5

const (
	NameKey MetricKey = iota
	GrpcEndpointKey
	ConfigNameKey
	PhaseKey
	Error
)

// mustInitialized is true when the field must be initialized
// before the reporter is started.
// the order must be the same as above
var mustInitialized = []bool{
	true,
	true,
	false,
	true,
	false,
}

type reporter struct {
	done     chan struct{}
	client   Client
	interval time.Duration
	// trigger a report immediately
	trigger chan struct{}

	// initialized is true when each of the
	// following fields is set at least once.
	initialized atomic.Bool
	// initmap is true when the field is set
	// at least once.
	initmap [filedNum]bool

	lock sync.Mutex
	data [filedNum]any
}

func NewReporter(client Client, interval time.Duration) Reporter {
	initmap := [filedNum]bool{}
	for i := range filedNum {
		initmap[i] = !mustInitialized[i]
	}

	return &reporter{
		done:     make(chan struct{}),
		client:   client,
		interval: interval,
		trigger:  make(chan struct{}, 1),
		initmap:  initmap,
	}
}

func (r *reporter) Close() {
	close(r.done)
}

func (r *reporter) Start() {
	timer := time.NewTimer(r.interval)

	for {
		select {
		case <-r.done:
			timer.Stop()
			return
		case <-r.trigger:
		case <-timer.C:
		}

		r.tryReport()

		timer = time.NewTimer(r.interval)
	}
}

func (r *reporter) tryReport() {
	// if not initialized, wait for initialization
	if !r.initialized.Load() {
		log.Info("Not all fields are set yet, waiting for initialization, no report will be sent now.")
		return
	}

	// report
	status := report.Status{}
	for i := range filedNum {
		if !r.initmap[i] {
			continue
		}

		v := r.data[i]
		if v == nil {
			continue
		}

		switch MetricKey(i) {
		case NameKey:
			status.Name = v.(string)
		case GrpcEndpointKey:
			status.GrpcEndpoint = v.(string)
		case ConfigNameKey:
			status.ConfigName = v.(string)
		case PhaseKey:
			status.Phase = v.(report.Phase)
		case Error:
			t := v.(ErrorTime)
			status.Error = &report.ErrorTime{
				Message: t.Message,
				RetryAt: t.RetryAt,
			}
		}
	}

	// set request timeout to 1/2 * interval
	timeout := r.interval / 2
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := r.client.Report(ctx, &status)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("report timeout after %s", timeout)
		}

		log.Error("report status to orch", log.ErrorField(err))
	}
}

func (r *reporter) SetData(key MetricKey, value any) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.setData(key, value)
}

func (r *reporter) SetDatas(key []MetricKey, value []any) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(key) != len(value) {
		panic(fmt.Sprintf("key and value length not equal, key: %d, value: %d", len(key), len(value)))
	}

	for i := range key {
		r.setData(key[i], value[i])
	}
}

func (r *reporter) setData(key MetricKey, value any) {
	if key < 0 || key >= filedNum {
		panic(fmt.Sprintf("MetricKey is %d out of range [0, %d)", key, filedNum))
	}
	r.data[key] = value
	r.checkInit(key)
}

func (r *reporter) checkInit(name MetricKey) {
	if !r.initialized.Load() {
		r.initmap[name] = true
		if !allTrue(r.initmap) {
			return
		}
		if !r.initialized.CompareAndSwap(false, true) {
			return
		}

		r.trigger <- struct{}{}
	}
}

func allTrue(arr [filedNum]bool) bool {
	for _, v := range arr {
		if !v {
			return false
		}
	}
	return true
}
