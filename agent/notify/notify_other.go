//go:build !linux && !windows

package notify

func Ready() error               { return nil }
func Reloading() error           { return nil }
func Stopping() error            { return nil }
func Status(_ string) error      { return nil }
func Error(_ error, _ int) error { return nil }
