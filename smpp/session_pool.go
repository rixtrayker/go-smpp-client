package smpp

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
)

type sessionPool struct {
	config   Config
	sessions []*gosmpp.Session
	mutex    sync.Mutex
}

func getSessionPool(maxSessions int, config Config) (*sessionPool, error) {
	pool := &sessionPool{
		config:   config,
		sessions: make([]*gosmpp.Session, 0, maxSessions),

	}

	for i := 0; i < maxSessions; i++ {
		session, err := pool.createSession()
		if err != nil {
			return nil, err
		}
		pool.sessions = append(pool.sessions, session)
	}

	return pool, nil
}

func (p *sessionPool) getSession() *gosmpp.Session {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.sessions) == 0 {
		return nil
	}

	session := p.sessions[0]
	p.sessions = p.sessions[1:]
	return session
}

func (p *sessionPool) returnSession(session *gosmpp.Session) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.sessions = append(p.sessions, session)
}

func (p *sessionPool) submitSMSToPool(msg string) error {
	session := p.getSession()
	if session == nil {
		return fmt.Errorf("no available sessions in the pool")
	}
	defer p.returnSession(session)

	handler, _ := NewSMPPHandler()
	if err := session.Transceiver().Submit(handler.newSubmitSM(msg)); err != nil {
		return err
	}

	return nil
}

func (p *sessionPool) createSession() (*gosmpp.Session, error) {
	auth := gosmpp.Auth{
		SMSC:       fmt.Sprintf("%s:%d", p.config.SMSCHost, p.config.SMSCPort),
		SystemID:   p.config.SystemID,
		Password:   p.config.Password,
		SystemType: "",
	}

	trans, err := gosmpp.NewSession(
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

			OnPDU: p.handlePDU(),

			OnClosed: func(state gosmpp.State) {
				fmt.Print("Closed connection, state: ")
				fmt.Println(state)
			},
		}, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return trans, nil
}


func (p *sessionPool) handlePDU() func(pdu.PDU, bool) {
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

// func Close
func (p *sessionPool) Close() {
	for _, session := range p.sessions {
		session.Close()
	}
}