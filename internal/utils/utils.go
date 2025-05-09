package utils

import (
	"log"
	"os"
	"os/signal"
)

func WaitSignal(signals []os.Signal, fns ...func()) {
	stopSigCh := make(chan os.Signal, 1)
	signal.Notify(stopSigCh, signals...)

	canStopCh := make(chan struct{})
	count := 0
	for {
		select {
		case sig := <-stopSigCh:
			count++
			log.Printf("Receive signal: %s, count: %d", sig.String(), count)

			if count == 1 {
				log.Println("First signal received, initiating graceful shutdown...")
				go func() {
					for _, fn := range fns {
						fn()
					}
					close(canStopCh)
				}()
			} else if count >= 3 {
				log.Println("Receive signal again, force exit.")
				os.Exit(1)
			}
		case <-canStopCh:
			return
		}
	}
}
