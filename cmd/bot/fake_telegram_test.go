package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
)

const (
	testToken  = "123456:abcdefghijklmnopqrstuvwxyz012345678"
	testChatID = int64(100)
	testUserID = int64(100)
)

// apiCall is one Bot API call received by fakeTelegram.
type apiCall struct {
	method string
	params map[string]any
}

// fakeTelegram is a telegoapi.Caller that records Bot API calls and answers them as Telegram would.
type fakeTelegram struct {
	mu    sync.Mutex
	calls []apiCall
	// sendErrors makes sendMessage to these chats fail with the given Bot API error code.
	sendErrors map[int64]int
}

func (f *fakeTelegram) Call(ctx context.Context, url string, data *telegoapi.RequestData) (*telegoapi.Response, error) {
	method := url[strings.LastIndex(url, "/")+1:]

	params, err := requestParams(data)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.calls = append(f.calls, apiCall{method: method, params: params})
	// Each sent message gets its call number as its ID, so messages can be told apart.
	messageID := len(f.calls)
	f.mu.Unlock()

	switch method {
	case "getMe":
		return okResponse(`{"id":1,"is_bot":true,"first_name":"Test","username":"test_bot"}`), nil
	case "getUpdates":
		// Long polling: wait until the bot is stopped.
		<-ctx.Done()

		return nil, ctx.Err()
	case "sendMessage":
		chatID := asInt64(params["chat_id"])

		if code, ok := f.sendErrors[chatID]; ok {
			return &telegoapi.Response{Ok: false, Error: &telegoapi.Error{ErrorCode: code, Description: "Forbidden"}}, nil
		}

		return okResponse(fmt.Sprintf(`{"message_id":%d,"date":0,"chat":{"id":%d,"type":"private"}}`, messageID, chatID)), nil
	default:
		return okResponse("true"), nil
	}
}

func requestParams(data *telegoapi.RequestData) (map[string]any, error) {
	params := make(map[string]any)

	if data == nil {
		return params, nil
	}

	body := data.BodyRaw
	if body == nil && data.BodyStream != nil {
		var err error

		if body, err = io.ReadAll(data.BodyStream); err != nil {
			return nil, err
		}
	}

	if len(body) == 0 {
		return params, nil
	}

	if err := json.Unmarshal(body, &params); err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}

	return params, nil
}

func okResponse(result string) *telegoapi.Response {
	return &telegoapi.Response{Ok: true, Result: json.RawMessage(result)}
}

// Decoded JSON helpers: they return the zero value when v has another type.

func asString(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	return s
}

func asInt64(v any) int64 {
	f, ok := v.(float64)
	if !ok {
		return 0
	}

	return int64(f)
}

func asMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	return m
}

func asSlice(v any) []any {
	s, ok := v.([]any)
	if !ok {
		return nil
	}

	return s
}

// callsOf returns the recorded calls of method.
func (f *fakeTelegram) callsOf(method string) []apiCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []apiCall

	for _, c := range f.calls {
		if c.method == method {
			out = append(out, c)
		}
	}

	return out
}

// texts returns the texts of all sent messages.
func (f *fakeTelegram) texts() []string {
	var out []string

	for _, c := range f.callsOf("sendMessage") {
		out = append(out, asString(c.params["text"]))
	}

	return out
}

// lastText returns the text of the last sent message.
func (f *fakeTelegram) lastText(t *testing.T) string {
	t.Helper()

	texts := f.texts()
	require.NotEmpty(t, texts, "no message was sent")

	return texts[len(texts)-1]
}

// lastKeyboardMessageID returns the ID of the last sent message with an inline keyboard, or 0.
func (f *fakeTelegram) lastKeyboardMessageID() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := len(f.calls) - 1; i >= 0; i-- {
		if f.calls[i].method == "sendMessage" && f.calls[i].params["reply_markup"] != nil {
			return i + 1
		}
	}

	return 0
}

// lastButtons returns the callback data of the last sent message's inline keyboard.
func (f *fakeTelegram) lastButtons(t *testing.T) []string {
	t.Helper()

	sent := f.callsOf("sendMessage")
	require.NotEmpty(t, sent, "no message was sent")

	markup := asMap(sent[len(sent)-1].params["reply_markup"])

	var out []string

	for _, row := range asSlice(markup["inline_keyboard"]) {
		for _, button := range asSlice(row) {
			out = append(out, asString(asMap(button)["callback_data"]))
		}
	}

	return out
}

func newFakeBot(t *testing.T, api *fakeTelegram) *telego.Bot {
	t.Helper()

	bot, err := telego.NewBot(testToken, telego.WithAPICaller(api), telego.WithDiscardLogger())
	require.NoError(t, err)

	return bot
}

// fakeService is a service.Service that records requests and returns preset results.
type fakeService struct {
	mu          sync.Mutex
	convertReqs []service.ConvertRequest
	calcReqs    []service.CalculateRequest
	convertResp *service.ConvertResponse
	calcResp    *service.CalculateResponse
	err         error
	panicMsg    string
}

func (s *fakeService) Convert(_ context.Context, p service.ConvertRequest) (*service.ConvertResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.convertReqs = append(s.convertReqs, p)

	return s.convertResp, s.err
}

func (s *fakeService) Calculate(_ context.Context, p service.CalculateRequest) (*service.CalculateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.panicMsg != "" {
		panic(s.panicMsg)
	}

	s.calcReqs = append(s.calcReqs, p)

	return s.calcResp, s.err
}

// botHarness runs the real handlers over a fake Telegram API and feeds them updates one at a time.
type botHarness struct {
	t       *testing.T
	api     *fakeTelegram
	store   *sessionStore
	users   *userStore
	updates chan telego.Update
	handled chan struct{}
	nextID  int
}

func newBotHarness(t *testing.T, svc service.Service) *botHarness {
	t.Helper()

	api := &fakeTelegram{}

	h := &botHarness{
		t:       t,
		api:     api,
		store:   newSessionStore(),
		users:   newTestStore(t),
		updates: make(chan telego.Update),
		handled: make(chan struct{}, 1),
	}

	var running inflight

	bh, err := telegohandler.NewBotHandler(newFakeBot(t, api), running.track(h.updates))
	require.NoError(t, err)

	// Signals that an update went through all handlers, so tests can check the result.
	bh.Use(func(ctx *telegohandler.Context, update telego.Update) error {
		defer func() { h.handled <- struct{}{} }()

		return ctx.Next(update)
	})
	bh.Use(running.done)
	registerHandlers(bh, h.store, svc, h.users)

	stopped := make(chan error, 1)

	go func() { stopped <- bh.Start() }()

	t.Cleanup(func() {
		close(h.updates)
		require.NoError(t, <-stopped)
	})

	return h
}

func (h *botHarness) send(update telego.Update) {
	h.t.Helper()

	h.nextID++
	update.UpdateID = h.nextID

	h.updates <- update

	select {
	case <-h.handled:
	case <-time.After(5 * time.Second):
		h.t.Fatal("the update was not handled")
	}
}

// text sends a text message (or a command) from the test user.
func (h *botHarness) text(text string) {
	h.t.Helper()

	h.send(telego.Update{Message: &telego.Message{
		MessageID: h.nextID + 1,
		Chat:      telego.Chat{ID: testChatID, Type: telego.ChatTypePrivate},
		From:      &telego.User{ID: testUserID, FirstName: "Test"},
		Text:      text,
	}})
}

// press presses a button with data on the latest keyboard, as the test user.
func (h *botHarness) press(data string) {
	h.t.Helper()

	h.pressOn(h.api.lastKeyboardMessageID(), data)
}

// pressOn presses a button with data on the message with messageID, as the test user.
func (h *botHarness) pressOn(messageID int, data string) {
	h.t.Helper()

	h.send(telego.Update{CallbackQuery: &telego.CallbackQuery{
		ID:      fmt.Sprintf("q%d", h.nextID+1),
		From:    telego.User{ID: testUserID, FirstName: "Test"},
		Message: &telego.Message{MessageID: messageID, Chat: telego.Chat{ID: testChatID, Type: telego.ChatTypePrivate}},
		Data:    data,
	}})
}
