package db

import (
	"sync"
	"time"

	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"

	"github.com/prometheus/client_golang/prometheus"
)

var tokenSessionTimeout = 30 * time.Minute
var tokenSessionsActive = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "suggestion_tokens_count",
	Help: "Number of active suggestion sessions",
})

var tokenSessionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "suggestion_tokens_total",
	Help: "Total suggestion sessions since restart",
})

type tokenData struct {
	ID        shared.SuggestionToken
	CreatedAt time.Time // When the token was created
	URL       string    // The url related to the token, used for quick validation.
}

type sessionTokenStore struct {
	sync.RWMutex
	sessions map[shared.SuggestionToken]tokenData
}

func (s *sessionTokenStore) Get(id shared.SuggestionToken) (tokenData, bool) {
	s.RLock()
	data, ok := s.sessions[id]
	s.RUnlock()
	return data, ok
}

func (s *sessionTokenStore) Reset(sessions []tokenData) {
	s.Lock()
	defer func() {
		s.Unlock()
		s.expireSessions()
	}()
	s.sessions = make(map[shared.SuggestionToken]tokenData, 0)
	for _, v := range sessions {
		s.sessions[v.ID] = v
	}
}

func (s *sessionTokenStore) New(URL string) shared.SuggestionToken {
	idstr, err := shared.SecureRandomString(32)
	if err != nil {
		panic(err)
	}
	id := shared.SuggestionToken(idstr)
	s.Lock()
	s.sessions[id] = tokenData{
		ID:        id,
		CreatedAt: time.Now(),
		URL:       URL,
	}
	s.Unlock()
	tokenSessionsActive.Add(1)
	tokenSessionsTotal.Add(1)
	return id
}

func (s *sessionTokenStore) expireSessions() {
	var expired []shared.SuggestionToken
	start := time.Now()
	th := start.Add(-tokenSessionTimeout)
	s.RLock()
	for _, v := range s.sessions {
		if v.CreatedAt.Before(th) {
			expired = append(expired, v.ID)
		}
	}
	s.RUnlock()
	if len(expired) > 0 {
		s.Lock()
		for _, v := range expired {
			delete(s.sessions, v)
		}
		tokenSessionsActive.Set(float64(len(s.sessions)))
		tokenSessionsTotal.Set(float64(len(s.sessions)))
		s.Unlock()
		if lg.V(3) {
			lg.Infof("expired %d sessions in %s", len(expired), time.Now().Sub(start).String())
		}
	}
}

func init() {
	prometheus.MustRegister(tokenSessionsActive)
	prometheus.MustRegister(tokenSessionsTotal)
	go func() {
		for {
			<-time.After(10 * time.Second)
			SessionTokens.expireSessions()
		}
	}()
}

var SessionTokens = sessionTokenStore{
	sessions: make(map[shared.SuggestionToken]tokenData, 0),
}
