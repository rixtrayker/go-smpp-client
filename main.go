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
	var wg sync.WaitGroup

	handler, err := smpp.NewSMPPHandler()
	if err != nil {
		fmt.Println("Error creating SMPP handler:", err)
		return
	}
	defer handler.Close()  // Ensure resources are cleaned up

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.SendAndReceiveSMS(ctx)
	}()

	go func() {
		<-quit
		fmt.Println("Shutting down...")
		cancel()
	}()

	wg.Wait()
}
