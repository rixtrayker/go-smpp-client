package main

import (
	"fmt"
	"sync"

	"github.com/rixtrayker/go-smpp/smpp"
)

func main() {
	var wg sync.WaitGroup


	handler, err := smpp.NewSMPPHandler()
	if err != nil {
		fmt.Println("Error creating SMPP handler:", err)
		return
	}

	wg.Add(1)
	go handler.SendAndReceiveSMS(&wg)

	wg.Wait()
}