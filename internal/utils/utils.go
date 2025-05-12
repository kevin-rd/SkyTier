package utils

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"
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

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
