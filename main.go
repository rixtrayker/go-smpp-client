package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rixtrayker/go-smpp/smpp"
)

func main() {
	config, err := smpp.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	sessionPool, err := smpp.NewSessionPool(config, 3)
	if err != nil {
		fmt.Println("Error creating session pool:", err)
		return
	}

	handler := smpp.NewSMPPHandler(config, sessionPool)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.SendAndReceiveSMS(ctx, &wg)
	}()

	go func() {
		<-quit
		fmt.Println("Shutting down...")
		handler.Close()
		os.Exit(1)
	}()

	wg.Wait()
}
