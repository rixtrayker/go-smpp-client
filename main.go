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

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		handler.SendAndReceiveSMS(ctx, &wg)
	}()

	go func() {
		<-quit
		println("Shutting down...")

		handler.Close()

		os.Exit(1)
	}()

	wg.Wait()
}
