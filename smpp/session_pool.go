package smpp

import (
	"fmt"
	"sync"

	"github.com/linxGnu/gosmpp/pdu"
)

// SessionPool manages a pool of SMPP sessions.
type SessionPool struct {
	config   Config
	sessions []*SingleSession
	mutex    sync.Mutex
}

// NewSessionPool creates a new session pool.
func NewSessionPool(config Config, maxSessions int) (*SessionPool, error) {
	pool := &SessionPool{
		config:   config,
		sessions: make([]*SingleSession, 0, maxSessions),
	}

	for i := 0; i < maxSessions; i++ {
		session, err := NewSingleSession(config)
		if err != nil {
			return nil, err
		}
		pool.sessions = append(pool.sessions, session)
	}

	return pool, nil
}

func (p *SessionPool) getSession() *SingleSession {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.sessions) == 0 {
		return nil
	}

	session := p.sessions[0]
	p.sessions = p.sessions[1:]
	return session
}

func (p *SessionPool) returnSession(session *SingleSession) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.sessions = append(p.sessions, session)
}

func (p *SessionPool) SubmitSM(submitSM *pdu.SubmitSM) error {
	session := p.getSession()
	if session == nil {
		return fmt.Errorf("no available sessions in the pool")
	}
	defer p.returnSession(session)

	return session.SubmitSM(submitSM)
}

func (p *SessionPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, session := range p.sessions {
		session.Close()
	}
}
