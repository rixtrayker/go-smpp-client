package smpp

import (
	"context"
	"fmt"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

type ISMPPHandler interface {
	SendSMS(ctx context.Context, msg string) error
	NewSubmitSM(msg string) *pdu.SubmitSM
	Close()
}

type SMPPHandler struct {
	config      Config
	session     Session
	rateLimiter *RateLimiter
}

func NewSMPPHandler(config Config, session Session, rateLimiter *RateLimiter) ISMPPHandler {
	return &SMPPHandler{
		config:      config,
		session:     session,
		rateLimiter: rateLimiter,
	}
}

func (h *SMPPHandler) SendSMS(ctx context.Context, msg string) error {
	submitSM := h.NewSubmitSM(msg)

	if s, ok := h.session.(*SingleSession); ok {
		if !h.rateLimiter.Allow(s) {
			return fmt.Errorf("rate limit exceeded, try again later")
		}
	}

	return h.session.SubmitSM(submitSM)
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
	h.session.Close()
}
