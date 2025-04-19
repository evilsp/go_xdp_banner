package statusfsm

import (
	"context"
	"errors"
	"xdp-banner/agent/internal/client"
	"xdp-banner/api/orch/v1/agent/report"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/queue"

	"github.com/looplab/fsm"
)

type Action = string

const (
	Start  Action = "start"
	Stop   Action = "stop"
	Reload Action = "reload"
)

type Event struct {
	Action Action
	Args   []any
}

func BeforeEventCallbackKey(action Action) string {
	return "before_" + action
}

type StatusFSM struct {
	queue queue.Queue[Event]
	fsm   *fsm.FSM
	done  chan struct{}
}

func New(startCallback fsm.Callback, stopCallback fsm.Callback, reloadCallback fsm.Callback) *StatusFSM {
	events := fsm.Events{
		{Name: Start, Src: []string{report.Phase_Ready.String(), report.Phase_Stopped.String()}, Dst: report.Phase_Running.String()},
		{Name: Stop, Src: []string{report.Phase_Running.String(), report.Phase_Ready.String()}, Dst: report.Phase_Stopped.String()},
		{Name: Reload, Src: []string{report.Phase_Running.String()}, Dst: report.Phase_Running.String()},
	}

	callback := fsm.Callbacks{
		BeforeEventCallbackKey(Start):  startCallback,
		BeforeEventCallbackKey(Stop):   stopCallback,
		BeforeEventCallbackKey(Reload): reloadCallback,
		"enter_state": func(_ context.Context, e *fsm.Event) {
			client.SetPhase(report.ToPhase(e.Dst))
		},
	}

	client.SetPhase(report.Phase_Ready)
	fsm := fsm.NewFSM(
		report.Phase_Ready.String(),
		events,
		callback,
	)

	sf := &StatusFSM{
		queue: queue.NewSafeQueue[Event](),
		fsm:   fsm,
		done:  make(chan struct{}),
	}

	go sf.run()

	return sf
}

func (sf *StatusFSM) run() {
	for {
		select {
		case <-sf.done:
			return
		default:
		}

		event := sf.queue.Pop()

		err := sf.fsm.Event(context.Background(), event.Action, event.Args...)
		if err != nil {
			// when fsm dst is the same as src, it will return NoTransitionError
			// but we allow this to happen,
			// for example, reload event which is from Running to Running
			if errors.As(err, &fsm.NoTransitionError{}) {
				return
			}
			log.Warn("agent control", log.StringField("action", event.Action), log.AnyField("phase", sf.Current()), log.ErrorField(err))
			client.SetError(err.Error())
		}
	}
}

func (sf *StatusFSM) Close() {
	close(sf.done)
}

func (sf *StatusFSM) Event(action Action, args ...any) {
	sf.queue.Push(Event{
		Action: action,
		Args:   args,
	})
}

func (sf *StatusFSM) Current() report.Phase {
	return report.ToPhase(sf.fsm.Current())
}
