// Package notify provides facilities for notifying process managers
// of state changes, mainly for when running as a system service.
package notify

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// The documentation about this IPC protocol is available here:
// https://www.freedesktop.org/software/systemd/man/sd_notify.html

func sdNotify(payload string) error {
	if socketPath == "" {
		return nil
	}

	socketAddr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(payload))
	return err
}

// Ready notifies systemd that caddy has finished its
// initialization routines.
func Ready() error {
	return sdNotify("READY=1")
}

// Reloading notifies systemd that caddy is reloading its config.
func Reloading() error {
	return sdNotify("RELOADING=1")
}

// Stopping notifies systemd that caddy is stopping.
func Stopping() error {
	return sdNotify("STOPPING=1")
}

// Status sends systemd an updated status message.
func Status(msg string) error {
	return sdNotify("STATUS=" + msg)
}

// Error is like Status, but sends systemd an error message
// instead, with an optional errno-style error number.
func Error(err error, errno int) error {
	collapsedErr := strings.ReplaceAll(err.Error(), "\n", " ")
	msg := fmt.Sprintf("STATUS=%s", collapsedErr)
	if errno > 0 {
		msg += fmt.Sprintf("\nERRNO=%d", errno)
	}
	return sdNotify(msg)
}

var socketPath, _ = os.LookupEnv("NOTIFY_SOCKET")
