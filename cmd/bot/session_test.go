package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

func TestSessionStepID(t *testing.T) {
	tests := []struct {
		name string
		sess *session
		want string
	}{
		{name: "no flow", sess: &session{}, want: stepWithoutKeyboard},
		{name: "calculate year", sess: &session{flow: flowCalculate, calcStep: calcStepYear}, want: stepCalcYear},
		{name: "calculate text step", sess: &session{flow: flowCalculate, calcStep: calcStepAmount}, want: stepWithoutKeyboard},
		{name: "convert target currency", sess: &session{flow: flowConvert, convertStep: convertStepCurrencyTo}, want: stepConvertTo},
		{name: "convert done", sess: &session{flow: flowConvert, convertStep: convertStepDone}, want: stepWithoutKeyboard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sess.stepID())
		})
	}
}

// TestSerializeMsg runs with -race in CI: handlers of one user mutate the session
// without their own locks, so they must never overlap.
func TestSerializeMsg(t *testing.T) {
	const (
		userID  = 42
		updates = 20
	)

	store := newSessionStore()

	var active, maxActive atomic.Int32

	h := serializeMsg(store, func(_ *telegohandler.Context, msg telego.Message) error {
		n := active.Add(1)
		defer active.Add(-1)

		for {
			prev := maxActive.Load()
			if n <= prev || maxActive.CompareAndSwap(prev, n) {
				break
			}
		}

		sess := store.get(msg.From.ID)
		sess.incomes = append(sess.incomes, sess.currentInc)

		time.Sleep(time.Millisecond)

		return nil
	})

	var wg sync.WaitGroup

	for range updates {
		wg.Go(func() {
			assert.NoError(t, h(nil, telego.Message{From: &telego.User{ID: userID}}))
		})
	}

	wg.Wait()

	assert.Equal(t, int32(1), maxActive.Load())
	assert.Len(t, store.get(userID).incomes, updates)
}

func TestSerializeMsgIgnoresMessageWithoutSender(t *testing.T) {
	called := false

	h := serializeMsg(newSessionStore(), func(_ *telegohandler.Context, _ telego.Message) error {
		called = true

		return errors.New("must not be called")
	})

	require.NoError(t, h(nil, telego.Message{From: nil}))
	assert.False(t, called)
}

func TestSessionStoreLockIsPerUser(t *testing.T) {
	store := newSessionStore()

	unlock1 := store.lock(1)
	defer unlock1()

	done := make(chan struct{})

	go func() {
		unlock2 := store.lock(2)
		unlock2()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("lock of user 2 waited for user 1")
	}
}

func TestUpdateChatID(t *testing.T) {
	const chatID = 7

	tests := []struct {
		name   string
		update telego.Update
		wantOK bool
	}{
		{
			name:   "message",
			update: telego.Update{Message: &telego.Message{Chat: telego.Chat{ID: chatID}}},
			wantOK: true,
		},
		{
			name: "callback",
			update: telego.Update{CallbackQuery: &telego.CallbackQuery{
				Message: &telego.Message{Chat: telego.Chat{ID: chatID}},
			}},
			wantOK: true,
		},
		{
			name:   "callback without message",
			update: telego.Update{CallbackQuery: &telego.CallbackQuery{}},
			wantOK: false,
		},
		{name: "other update", update: telego.Update{}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := updateChatID(tt.update)

			assert.Equal(t, tt.wantOK, ok)

			if tt.wantOK {
				assert.Equal(t, int64(chatID), got)
			}
		})
	}
}

func TestInflightWaitsForHandledUpdates(t *testing.T) {
	var running inflight

	updates := make(chan telego.Update)
	out := running.track(updates)

	// Both channels are unbuffered: receive each update, as the bot handler does, before sending the next.
	updates <- telego.Update{UpdateID: 1}

	<-out

	updates <- telego.Update{UpdateID: 2}

	<-out

	// Finish only the first update.
	running.wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	assert.False(t, running.wait(ctx), "one update is still being handled")

	running.wg.Done()
	close(updates)

	assert.True(t, running.wait(context.Background()))

	_, open := <-out
	assert.False(t, open, "the forwarded channel closes with the source")
}

func TestSessionRestartIncomesKeepsTaxDetails(t *testing.T) {
	first := service.Income{DateRequest: service.DateRequest{Year: "2025", Month: "June", Day: "08"}, Currency: "USD", Amount: "100"}

	sess := &session{
		flow:       flowCalculate,
		calcStep:   calcStepConfirm,
		calcReq:    service.CalculateRequest{TaxType: "Small Business", YearIncome: "1000", Income: []service.Income{first}},
		incomes:    []service.Income{first},
		currentInc: first,
	}

	sess.restartIncomes()

	assert.Equal(t, calcStepYear, sess.calcStep)
	assert.Equal(t, flowCalculate, sess.flow)
	assert.Empty(t, sess.incomes)
	assert.Empty(t, sess.calcReq.Income)
	assert.Equal(t, service.Income{}, sess.currentInc)
	assert.Equal(t, "Small Business", sess.calcReq.TaxType)
	assert.Equal(t, "1000", sess.calcReq.YearIncome)

	opts, _, err := stepOptions(sess)
	require.NoError(t, err)
	assert.Greater(t, len(opts), 1, "any year can be chosen again after the restart")
}
