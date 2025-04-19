package main

import (
	"log"
	"xdp-banner/agent/cmd"
)

func main() {
	if err := cmd.NewAgentCmd().Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
