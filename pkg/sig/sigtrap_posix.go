// Copyright 2015 Matthew Holt and The Caddy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows && !plan9 && !nacl && !js

package sig

import (
	"os"
	"os/signal"
	"syscall"
	"xdp-banner/pkg/log"

	"go.uber.org/zap"
)

// trapSignalsPosix captures POSIX-only signals.
func trapSignalsPosix(exitProcess func()) {
	// Ignore all SIGPIPE signals to prevent weird issues with systemd: https://github.com/dunglas/frankenphp/issues/1020
	// Docker/Moby has a similar hack: https://github.com/moby/moby/blob/d828b032a87606ae34267e349bf7f7ccb1f6495a/cmd/dockerd/docker.go#L87-L90
	signal.Ignore(syscall.SIGPIPE)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

		for sig := range sigchan {
			switch sig {
			case syscall.SIGQUIT:
				log.Info("quitting process immediately", zap.String("signal", "SIGQUIT"))
				os.Exit(ExitCodeForceQuit)

			case syscall.SIGTERM:
				log.Info("shutting down apps, then terminating", zap.String("signal", "SIGTERM"))
				go exitProcess()

			case syscall.SIGUSR1:
				log.Info("not implemented", zap.String("signal", "SIGUSR1"))

			case syscall.SIGUSR2:
				log.Info("not implemented", zap.String("signal", "SIGUSR2"))

			case syscall.SIGHUP:
				// ignore; this signal is sometimes sent outside of the user's control
				log.Info("not implemented", zap.String("signal", "SIGHUP"))
			}
		}
	}()
}
