package smpp

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

type SMPPHandler struct {
	config     Config
	sessionPoo *sessionPool
}

func NewSMPPHandler() (*SMPPHandler, error) {
	config, err := LoadConfig()

	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil, err
	}

	pool, err := getSessionPool(10, config)
	if err != nil {
		return nil, err
	}

	return &SMPPHandler{
		config:     config,
		sessionPoo: pool,
	}, nil
}

func (h *SMPPHandler) SendAndReceiveSMS(wg *sync.WaitGroup) {
	defer wg.Done()

	// sending SMS(s)
	for i := 0; i < 1800; i++ {
		msg := fmt.Sprintf("MSG %d", i)
		if err := h.sessionPoo.submitSMSToPool(msg); err != nil {
			fmt.Println(err)
		}
	}
}

func (h *SMPPHandler) handlePDU() func(pdu.PDU, bool) {
	concatenated := map[uint8][]string{}
	return func(p pdu.PDU, _ bool) {
		switch pd := p.(type) {
		case *pdu.SubmitSMResp:
			fmt.Printf("SubmitSMResp:")

		case *pdu.GenericNack:
			fmt.Println("GenericNack Received")

		case *pdu.EnquireLinkResp:
			fmt.Println("EnquireLinkResp Received")

		case *pdu.DataSM:
			fmt.Printf("DataSM:")

		case *pdu.DeliverSM:
			fmt.Printf("DeliverSM:")
			// log.Println(pd.Message.GetMessage())
			// region concatenated sms (sample code)
			message, err := pd.Message.GetMessage()
			if err != nil {
				log.Fatal(err)
			}
			totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
			if found {
				if _, ok := concatenated[reference]; !ok {
					concatenated[reference] = make([]string, totalParts)
				}
				concatenated[reference][sequence-1] = message
			}
			if !found {
				log.Println(message)
			} else if parts, ok := concatenated[reference]; ok && isConcatenatedDone(parts, totalParts) {
				fmt.Println(strings.Join(parts, ""))
				// print byte reference
				// fmt.Println("reference: %s", string(reference))
				delete(concatenated, reference)
			}
			// endregion
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