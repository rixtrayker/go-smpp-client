package smpp

import (
	"context"
	"fmt"
	"sync"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

type SMPPHandler struct {
	config     Config
	sessionPool *sessionPool
}

func NewSMPPHandler() (*SMPPHandler, error) {
	config, err := LoadConfig()

	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil, err
	}

	pool, err := getSessionPool(3, config)
	if err != nil {
		return nil, err
	}

	return &SMPPHandler{
		config:     config,
		sessionPool: pool,
	}, nil
}

func (h *SMPPHandler) SendAndReceiveSMS(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()

    for i := 0; i < 1800; {
        select {
        case <-ctx.Done():
            fmt.Println("Context canceled, stopping SMS sending...")
            return // Exit the loop on context cancellation
        default:
            msg := fmt.Sprintf("MSG %d", i)
            if err := h.sessionPool.submitSMSToPool(msg); err != nil {
                fmt.Println(err)
            }
            i++
        }
    }
}



func (h *SMPPHandler) newSubmitSM(msg string) *pdu.SubmitSM {
	// build up submitSM
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(5)
	srcAddr.SetNpi(0)
	_ = srcAddr.SetAddress("00" + "522241")

	destAddr := pdu.NewAddress()
	destAddr.SetTon(1)
	destAddr.SetNpi(1)
	_ = destAddr.SetAddress("99" + "522241")

	submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
	submitSM.SourceAddr = srcAddr
	submitSM.DestAddr = destAddr
	_ = submitSM.Message.SetMessageWithEncoding(msg, data.UCS2)
	submitSM.ProtocolID = 0
	submitSM.RegisteredDelivery = 1
	submitSM.ReplaceIfPresentFlag = 0
	submitSM.EsmClass = 0

	return submitSM
}

func isConcatenatedDone(parts []string, total byte) bool {
	for _, part := range parts {
		if part != "" {
			total--
		}
	}
	return total == 0
}

// func Close

func (h *SMPPHandler) Close() {
	h.sessionPool.Close()
}