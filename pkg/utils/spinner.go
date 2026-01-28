package utils

import (
	"fmt"
	"time"
)

var (
	spinner       = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	tickerSpinner = time.NewTicker(time.Second / 10)
)

// RunWithSpinner runs a command with a spinner animation until it completes
// stop: channel to signal the spinner to stop
// finished: channel that will be closed when the spinner has finished cleaning up
func RunWithSpinner(message string, stop chan bool, finished chan bool) {
	tickerCounter := 0
	for {
		select {
		case <-stop:
			fmt.Println("\r✅", message, ": DONE.")
			close(finished)
			return
		case <-tickerSpinner.C:
			fmt.Printf("\r%s %s...", spinner[tickerCounter%len(spinner)], message)
			tickerCounter++
		}
	}
}
