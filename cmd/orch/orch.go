package main

import (
	"xdp-banner/orch/cmd"
	"xdp-banner/pkg/log"
)

func main() {
	if err := cmd.NewOrchCmd().Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
