package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	wg.Add(1)

	go handler.SendAndReceiveSMS(ctx, &wg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
	wg.Wait()

	handler.Close()
	fmt.Println("Shutting down...")
}
