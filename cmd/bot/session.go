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

// Step IDs are sent in callback data (cb:<step>:<value>), so a button is accepted
// only by the step that sent its keyboard.
const (
	stepCalcTaxType     = "calc_tax"
	stepCalcYear        = "calc_year"
	stepCalcMonth       = "calc_month"
	stepCalcDay         = "calc_day"
	stepCalcCurrency    = "calc_cur"
	stepCalcAddMore     = "calc_more"
	stepCalcConfirm     = "calc_ok"
	stepConvertYear     = "conv_year"
	stepConvertMonth    = "conv_month"
	stepConvertDay      = "conv_day"
	stepConvertFrom     = "conv_from"
	stepConvertTo       = "conv_to"
	stepConvertConfirm  = "conv_ok"
	stepWithoutKeyboard = ""
)

var calcStepIDs = map[calcStep]string{
	calcStepTaxType:  stepCalcTaxType,
	calcStepYear:     stepCalcYear,
	calcStepMonth:    stepCalcMonth,
	calcStepDay:      stepCalcDay,
	calcStepCurrency: stepCalcCurrency,
	calcStepAddMore:  stepCalcAddMore,
	calcStepConfirm:  stepCalcConfirm,
}

var convertStepIDs = map[convertStep]string{
	convertStepYear:         stepConvertYear,
	convertStepMonth:        stepConvertMonth,
	convertStepDay:          stepConvertDay,
	convertStepCurrencyFrom: stepConvertFrom,
	convertStepCurrencyTo:   stepConvertTo,
	convertStepConfirm:      stepConvertConfirm,
}

// session holds per-user conversation state.
type session struct {
	flow flowType

	// tax calculation state
	calcStep   calcStep
	calcReq    service.CalculateRequest
	currentInc service.Income
	incomes    []service.Income

	// convert state
	convertStep convertStep
	convertReq  service.ConvertRequest
}

// stepID returns the ID of the current step, or stepWithoutKeyboard when the step expects text.
func (s *session) stepID() string {
	switch s.flow {
	case flowCalculate:
		return calcStepIDs[s.calcStep]
	case flowConvert:
		return convertStepIDs[s.convertStep]
	case flowNone:
	}

	return stepWithoutKeyboard
}

// restartIncomes clears the captured incomes and returns to the first income's year step,
// keeping the tax type and the year income, as the CLI does on "No" at the final review.
func (s *session) restartIncomes() {
	s.incomes = nil
	s.currentInc = service.Income{}
	s.calcReq.Income = nil
	s.calcStep = calcStepYear
}

// sessionStore manages sessions for all users.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[int64]*session
	// locks serialize updates of one user; they are kept apart from sessions because reset replaces a session.
	locks map[int64]*sync.Mutex
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		sessions: make(map[int64]*session),
		locks:    make(map[int64]*sync.Mutex),
	}
}

// lock blocks until no other update of userID is being handled and returns the unlock func.
func (s *sessionStore) lock(userID int64) func() {
	s.mu.Lock()

	l, ok := s.locks[userID]
	if !ok {
		l = &sync.Mutex{}
		s.locks[userID] = l
	}

	s.mu.Unlock()

	l.Lock()

	return l.Unlock
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
