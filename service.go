// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// runService wait for os interrupt
func runService() chan bool {
	waitChan := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range c {
			waitChan <- true
			return
		}
	}()
	return waitChan
}
