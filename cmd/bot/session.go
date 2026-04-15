package main

import (
	"sync"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

// flowType represents which flow the user is currently in.
type flowType int

const (
	flowNone      flowType = iota
	flowCalculate          // tax calculation flow
	flowConvert            // currency conversion flow
)

// calcStep represents the step in the tax calculation flow.
type calcStep int

const (
	calcStepTaxType    calcStep = iota // select tax type
	calcStepYearIncome                 // enter year income in GEL
	calcStepYear                       // select income year
	calcStepMonth                      // select income month
	calcStepDay                        // select income day
	calcStepAmount                     // enter income amount
	calcStepCurrency                   // select income currency
	calcStepAddMore                    // ask to add another income
	calcStepConfirm                    // confirm all inputs
	calcStepDone                       // flow complete
)

// convertStep represents the step in the currency conversion flow.
type convertStep int

const (
	convertStepYear         convertStep = iota // select year
	convertStepMonth                           // select month
	convertStepDay                             // select day
	convertStepAmount                          // enter amount
	convertStepCurrencyFrom                    // select currency from
	convertStepCurrencyTo                      // select currency to
	convertStepConfirm                         // confirm all inputs
	convertStepDone                            // flow complete
)

// session holds per-user conversation state.
type session struct {
	flow flowType

	// tax calculation state
	calcStep    calcStep
	calcReq     service.CalculateRequest
	currentInc  service.Income
	incomes     []service.Income

	// convert state
	convertStep convertStep
	convertReq  service.ConvertRequest
}

// sessionStore manages sessions for all users.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[int64]*session
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[int64]*session),
	}
}

// get returns the session for the given userID, creating one if it doesn't exist.
func (s *sessionStore) get(userID int64) *session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[userID]
	if !ok {
		sess = &session{}
		s.sessions[userID] = sess
	}

	return sess
}

// reset clears the session for the given userID.
func (s *sessionStore) reset(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[userID] = &session{}
}
