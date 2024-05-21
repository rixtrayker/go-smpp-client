package smpp

import (
	"fmt"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
)

const maxOutstandingRequests = 100
// ISessionPool defines the interface for the session pool
type ISessionPool interface {
	SubmitSMSToPool(msg string) error
	Close()
}

// SessionPool implements ISessionPool
type SessionPool struct {
	config         Config
	sessions       []*gosmpp.Session
	mutex          sync.Mutex
	submitRespCh   chan struct{}
	waitBlockingCh chan struct{}
}

func NewSessionPool(config Config, maxSessions int) (ISessionPool, error) {
	pool := &SessionPool{
		config:         config,
		sessions:       make([]*gosmpp.Session, 0, maxSessions),
		submitRespCh:   make(chan struct{}, maxOutstandingRequests),
		waitBlockingCh: make(chan struct{}),
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

func (p *SessionPool) getSession() *gosmpp.Session {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.sessions) == 0 {
		return nil
	}

	session := p.sessions[0]
	p.sessions = p.sessions[1:]
	return session
}

func (p *SessionPool) returnSession(session *gosmpp.Session) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.sessions = append(p.sessions, session)
}

func (p *SessionPool) SubmitSMSToPool(msg string) error {
	p.mutex.Lock()
	if len(p.sessions) == 0 {
		p.mutex.Unlock()
		return fmt.Errorf("no available sessions in the pool")
	}

	session := p.sessions[0]
	p.sessions = p.sessions[1:]
	p.mutex.Unlock()

	defer func() {
		p.returnSession(session)
	}()

	handler := NewSMPPHandler(p.config, p)
	submitSM := handler.newSubmitSM(msg)

	if len(p.submitRespCh) == maxOutstandingRequests {
		<-p.waitBlockingCh
		time.Sleep(1 * time.Second)
	}

	if err := session.Transceiver().Submit(submitSM); err != nil {
		return err
	}

	p.submitRespCh <- struct{}{}

	return nil
}

func (p *SessionPool) createSession() (*gosmpp.Session, error) {
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

func (p *SessionPool) handlePDU() func(pdu.PDU, bool) {
	return func(pd pdu.PDU, _ bool) {
		switch pd.(type) {
		case *pdu.SubmitSMResp:
			fmt.Printf("SubmitSMResp received\n")
			<-p.submitRespCh
		case *pdu.GenericNack:
			fmt.Println("GenericNack Received")
		case *pdu.EnquireLinkResp:
			fmt.Println("EnquireLinkResp Received")
		case *pdu.DataSM:
			fmt.Printf("DataSM received\n")
		case *pdu.DeliverSM:
			fmt.Printf("DeliverSM received\n")
			// Extract message content from pdu.Message if needed
		}
	}
}

// Close closes all sessions in the pool.
func (p *SessionPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, session := range p.sessions {
		session.Close()
	}
}
