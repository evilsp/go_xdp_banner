package notify

import "golang.org/x/sys/windows/svc"

// globalStatus store windows service status, it can be
// use to notify caddy status.
var globalStatus chan<- svc.Status

func SetGlobalStatus(status chan<- svc.Status) {
	globalStatus = status
}

func Ready() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{
			State:   svc.Running,
			Accepts: svc.AcceptStop | svc.AcceptShutdown,
		}
	}
	return nil
}

func Reloading() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{State: svc.StartPending}
	}
	return nil
}

func Stopping() error {
	if globalStatus != nil {
		globalStatus <- svc.Status{State: svc.StopPending}
	}
	return nil
}

// TODO: not implemented
func Status(_ string) error { return nil }

// TODO: not implemented
func Error(_ error, _ int) error { return nil }
