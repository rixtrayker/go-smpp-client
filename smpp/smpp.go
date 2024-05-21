package smpp

import (
	"context"
	"fmt"
	"sync"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

// ISMPPHandler defines the interface for the SMPP handler.
type ISMPPHandler interface {
	SendAndReceiveSMS(ctx context.Context, wg *sync.WaitGroup)
	NewSubmitSM(msg string) *pdu.SubmitSM
	Close()
}

// SMPPHandler implements ISMPPHandler
type SMPPHandler struct {
	config      Config
	sessionPool ISessionPool
}

// NewSMPPHandler creates a new SMPPHandler
func NewSMPPHandler(config Config, sessionPool ISessionPool) ISMPPHandler {
	return &SMPPHandler{
		config:      config,
		sessionPool: sessionPool,
	}
}

func (h *SMPPHandler) SendAndReceiveSMS(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 1800; i++ {
		msg := fmt.Sprintf("MSG %d", i)
		if err := h.sessionPool.SubmitSMSToPool(h, msg); err != nil {
			fmt.Println("Error submitting SMS to pool:", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("Context canceled, shutting down...")
			return
		default:
		}
	}
}

func (h *SMPPHandler) NewSubmitSM(msg string) *pdu.SubmitSM {
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(5)
	srcAddr.SetNpi(0)
	_ = srcAddr.SetAddress("00522241")

	destAddr := pdu.NewAddress()
	destAddr.SetTon(1)
	destAddr.SetNpi(1)
	_ = destAddr.SetAddress("99522241")

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

func (h *SMPPHandler) Close() {
	h.sessionPool.Close()
}
