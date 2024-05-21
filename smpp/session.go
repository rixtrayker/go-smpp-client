package smpp

import (
	"fmt"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
)

// Session interface represents an SMPP session.
type Session interface {
	SubmitSM(submitSM *pdu.SubmitSM) error
	Close()
}

// SingleSession represents a single SMPP session.
type SingleSession struct {
	session *gosmpp.Session
}

// NewSingleSession creates a new single session.
func NewSingleSession(config Config) (*SingleSession, error) {
	auth := gosmpp.Auth{
		SMSC:       fmt.Sprintf("%s:%d", config.SMSCHost, config.SMSCPort),
		SystemID:   config.SystemID,
		Password:   config.Password,
		SystemType: "",
	}

	session, err := gosmpp.NewSession(
		gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
		gosmpp.Settings{
			EnquireLink: 2000 * time.Millisecond,
			ReadTimeout: 10 * time.Second,
			OnSubmitError: func(_ pdu.PDU, err error) {
				fmt.Println("SubmitPDU error:", err)
			},
			OnReceivingError: func(err error) {
				fmt.Println("Receiving PDU/Network error:", err)
			},
			OnRebindingError: func(err error) {
				fmt.Println("Rebinding but error:", err)
			},
			OnPDU: func(pd pdu.PDU, _ bool) {
				switch pd.(type) {
				case *pdu.SubmitSMResp:
					fmt.Printf("SubmitSMResp received\n")
				case *pdu.GenericNack:
					fmt.Println("GenericNack Received")
				case *pdu.EnquireLinkResp:
					fmt.Println("EnquireLinkResp Received")
				case *pdu.DataSM:
					fmt.Printf("DataSM received\n")
				case *pdu.DeliverSM:
					fmt.Printf("DeliverSM received\n")
				}
			},
			OnClosed: func(state gosmpp.State) {
				fmt.Print("Closed connection, state: ")
				fmt.Println(state)
			},
		}, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &SingleSession{session: session}, nil
}

func (s *SingleSession) SubmitSM(submitSM *pdu.SubmitSM) error {
	return s.session.Transceiver().Submit(submitSM)
}

func (s *SingleSession) Close() {
	s.session.Close()
}