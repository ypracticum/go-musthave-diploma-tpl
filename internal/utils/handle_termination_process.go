package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func HandleTerminationProcess(cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
}
