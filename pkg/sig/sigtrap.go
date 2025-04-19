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

package sig

import (
	"os"
	"os/signal"
	"xdp-banner/pkg/log"

	"go.uber.org/zap"
)

// TrapSignals create signal/interrupt handlers as best it can for the
// current OS. This is a rather invasive function to call in a Go program
// that captures signals already, so in that case it would be better to
// implement these handlers yourself.
func TrapSignals(exitProcess func()) {
	trapSignalsCrossPlatform(exitProcess)
	trapSignalsPosix()
}

// trapSignalsCrossPlatform captures SIGINT or interrupt (depending
// on the OS), which initiates a graceful shutdown. A second SIGINT
// or interrupt will forcefully exit the process immediately.
func trapSignalsCrossPlatform(exitProcess func()) {
	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt)

		for i := 0; true; i++ {
			<-shutdown

			if i > 0 {
				log.Warn("force quit", zap.String("signal", "SIGINT"))
				os.Exit(ExitCodeForceQuit)
			}

			log.Info("shutting down", zap.String("signal", "SIGINT"))
			go exitProcess()
		}
	}()
}

// Exit codes. Generally, you should NOT
// automatically restart the process if the
// exit code is ExitCodeFailedStartup (1).
const (
	ExitCodeSuccess = iota
	ExitCodeFailedStartup
	ExitCodeForceQuit
	ExitCodeFailedQuit
)
