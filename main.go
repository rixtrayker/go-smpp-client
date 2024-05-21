package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rixtrayker/go-smpp-client/smpp"
)

func main() {
	config, err := smpp.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	rateLimiter := smpp.NewRateLimiter(100, 20) // example limits
	rateLimiter.Reset()

	sessionPool, err := smpp.NewSessionPool(config, 10)
	if err != nil {
		fmt.Println("Error creating session pool:", err)
		return
	}

	handler := smpp.NewSMPPHandler(config, sessionPool, rateLimiter)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
		wg.Wait()
		handler.Close()
		fmt.Println("Shutting down...")
		os.Exit(0)
	}()

	// Call a function to test sending SMS
	testSendSMS(ctx, handler, &wg)

	wg.Wait()
}

func testSendSMS(ctx context.Context, handler smpp.ISMPPHandler, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1800; i++ {
			msg := fmt.Sprintf("MSG %d", i)
			err := handler.SendSMS(ctx, msg)
			if err != nil {
				fmt.Println("Error sending SMS:", err)
			}

			select {
			case <-ctx.Done():
				fmt.Println("Context canceled, stopping SMS sending...")
				return
			default:
				time.Sleep(100 * time.Millisecond) // simulate some delay between messages
			}
		}
	}()
}
