package main

import (
	"errors"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestBotServiceCommands(t *testing.T) {
	h := newBotHarness(t, &fakeService{})

	h.text("/start")
	assert.Contains(t, h.api.lastText(t), "Welcome to Georgia Tax Calculator Bot")

	h.text("/help")
	assert.Contains(t, h.api.lastText(t), "/calculate — Start tax calculation flow")

	h.text("/convert")
	require.Equal(t, flowConvert, h.store.get(testUserID).flow)

	h.text("/cancel")
	assert.Contains(t, h.api.lastText(t), "Use /calculate or /convert to start again")
	assert.Equal(t, flowNone, h.store.get(testUserID).flow)

	h.text("hello")
	assert.Equal(t, "Please use /calculate or /convert to start.", h.api.lastText(t))

	h.press("cb:conv_year:2024")
	assert.Equal(t, "Please use /calculate or /convert to start.", h.api.lastText(t))
	assert.Len(t, h.api.callsOf("answerCallbackQuery"), 1, "every callback query is answered")

	assert.Equal(t, []int64{testChatID}, h.users.All(), "the chat is tracked")
}

func TestBotConvertFlow(t *testing.T) {
	svc := &fakeService{convertResp: &service.ConvertResponse{
		Date:      time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC),
		Amount:    models.NewMoney(100, currencies.USD),
		Converted: models.NewMoney(270, currencies.GEL),
		Rate:      models.NewMoney(2.7, ""),
	}}
	h := newBotHarness(t, svc)

	h.text("/convert")
	assert.Contains(t, h.api.lastButtons(t), "cb:conv_year:2024")

	h.press("cb:conv_year:2024")
	assert.Contains(t, h.api.lastButtons(t), "cb:conv_month:March")

	h.press("cb:conv_month:March")
	assert.Contains(t, h.api.lastButtons(t), "cb:conv_day:15")

	h.press("cb:conv_day:15")
	assert.Contains(t, h.api.lastText(t), "Enter the amount to convert")

	h.press("cb:conv_day:16")
	assert.Contains(t, h.api.lastText(t), "Unexpected input", "the amount step expects text")

	h.text("12abc")
	assert.Contains(t, h.api.lastText(t), "❌ Invalid amount")

	h.text("100")
	assert.Contains(t, h.api.lastButtons(t), "cb:conv_from:"+currencies.USD)

	h.text("USD")
	assert.Contains(t, h.api.lastText(t), "Unexpected input", "the currency step expects a button")

	h.press("cb:conv_year:2024")
	assert.Contains(t, h.api.lastText(t), "Unexpected input", "a button of an earlier step is rejected")

	h.press("cb:conv_from:" + currencies.USD)
	h.press("cb:conv_to:" + currencies.GEL)
	assert.Contains(t, h.api.lastText(t), "Review your inputs")
	assert.Equal(t, []string{"cb:conv_ok:yes", "cb:conv_ok:no"}, h.api.lastButtons(t))

	h.press("cb:conv_ok:yes")

	texts := h.api.texts()
	assert.Contains(t, texts[len(texts)-2], "⏳ Converting...")
	assert.Contains(t, h.api.lastText(t), "Converted: 270 GEL")

	require.Len(t, svc.convertReqs, 1)
	assert.Equal(t, service.ConvertRequest{
		Year: "2024", Month: "March", Day: "15",
		CurrencyFrom: currencies.USD,
		CurrencyTo:   currencies.GEL,
		Amount:       "100",
	}, svc.convertReqs[0])
	assert.Equal(t, flowNone, h.store.get(testUserID).flow)
}

func TestBotConvertDeclinedAndFailed(t *testing.T) {
	h := newBotHarness(t, &fakeService{err: errors.New("NBG is down")})

	enter := func() {
		h.text("/convert")
		h.press("cb:conv_year:2024")
		h.press("cb:conv_month:March")
		h.press("cb:conv_day:15")
		h.text("100")
		h.press("cb:conv_from:" + currencies.EUR)
		h.press("cb:conv_to:" + currencies.GEL)
	}

	enter()
	h.press("cb:conv_ok:no")
	assert.Contains(t, h.api.lastText(t), "Restarting")
	assert.Contains(t, h.api.lastButtons(t), "cb:conv_year:2024")

	enter()
	h.press("cb:conv_ok:yes")
	assert.Contains(t, h.api.lastText(t), "❌ Conversion error: NBG is down")
}

func TestBotCalculateFlow(t *testing.T) {
	svc := &fakeService{calcResp: &service.CalculateResponse{
		TaxRate:              taxes.TaxRate{Type: taxes.TaxTypeSmallBusiness, Rate: 0.01},
		YearIncome:           models.NewMoney(270, currencies.GEL),
		TotalIncomeConverted: models.NewMoney(270, currencies.GEL),
		Tax:                  models.NewMoney(2.7, currencies.GEL),
		Declaration:          testDeclaration(),
	}}
	h := newBotHarness(t, svc)

	h.text("/calculate")
	assert.Contains(t, h.api.lastText(t), service.DeclarationsURL)
	assert.Contains(t, h.api.lastButtons(t), "cb:calc_tax:"+taxes.TaxTypeSmallBusiness.String())

	h.press("cb:calc_tax:" + taxes.TaxTypeSmallBusiness.String())
	assert.Contains(t, h.api.lastText(t),
		"💰 Enter field (15) of your previous month's declaration: income since the start of the year in GEL:\n"+
			"(0.00 for your first declaration of the year)")

	h.text("-1")
	assert.Contains(t, h.api.lastText(t), "❌ Invalid amount")

	h.text("0")

	addIncome := func(amount string) {
		h.press("cb:calc_year:2024")
		h.press("cb:calc_month:March")
		h.press("cb:calc_day:15")
		h.text(amount)
		h.press("cb:calc_cur:" + currencies.USD)
		h.press("cb:calc_inc_type:" + taxes.IncomeTypeNonCash.String())
	}

	h.press("cb:calc_year:2024")
	h.press("cb:calc_month:March")
	h.press("cb:calc_day:15")
	h.text("100")
	h.press("cb:calc_cur:" + currencies.USD)
	assert.Contains(t, h.api.lastText(t), "🧾 Select the income type (declaration field):\n• Cash (18) — ")
	assert.Equal(t, []string{
		"cb:calc_inc_type:Cash",
		"cb:calc_inc_type:POS terminal",
		"cb:calc_inc_type:Non-cash",
		"cb:calc_inc_type:Other",
		"cb:calc_inc_type:Unpaid supply",
	}, h.api.lastButtons(t))

	h.text("Non-cash")
	assert.Contains(t, h.api.lastText(t), "Unexpected input", "the type is chosen with a button")

	h.press("cb:calc_inc_type:" + taxes.IncomeTypeNonCash.String())
	assert.Contains(t, h.api.lastText(t), "✅ Income type: Non-cash (20)")
	assert.Contains(t, h.api.lastText(t), "  1) 2024-March-15 — 100 USD — Non-cash (20)\n")
	assert.Contains(t, h.api.lastText(t), "Add another income entry?")

	h.press("cb:calc_more:yes")
	assert.Equal(t, []string{"cb:calc_year:2024"}, h.api.lastButtons(t), "later incomes use the first income's year")

	h.press("cb:calc_year:2023")
	assert.Contains(t, h.api.lastText(t), "Unexpected input")

	addIncome("50")
	h.press("cb:calc_more:no")
	assert.Contains(t, h.api.lastText(t), "Review your inputs")
	assert.Contains(t, h.api.lastText(t), "  2) 2024-March-15 — 50 USD — Non-cash (20)\n")

	h.press("cb:calc_ok:no")
	assert.Contains(t, h.api.lastText(t), "Restarting income entry")
	assert.Equal(t, taxes.TaxTypeSmallBusiness.String(), h.store.get(testUserID).calcReq.TaxType, "the tax type is kept")

	addIncome("100")
	h.press("cb:calc_more:no")
	h.press("cb:calc_ok:yes")

	texts := h.api.texts()
	assert.Contains(t, texts[len(texts)-2], "⏳ Calculating taxes...")
	assert.Contains(t, h.api.lastText(t), "Taxes to Pay: 2.7 GEL")

	sent := h.api.callsOf("sendMessage")
	result := sent[len(sent)-1].params

	assert.Nil(t, result["parse_mode"], "the result is plain text with entities")

	entities := asSlice(result["entities"])
	require.Len(t, entities, 8)

	for _, e := range entities {
		assert.Equal(t, telego.EntityTypeCode, asString(asMap(e)["type"]))
	}

	require.Len(t, svc.calcReqs, 1)
	assert.Equal(t, service.CalculateRequest{
		Income: []service.Income{{
			Year: "2024", Month: "March", Day: "15",
			Currency: currencies.USD,
			Amount:   "100",
			Type:     taxes.IncomeTypeNonCash.String(),
		}},
		TaxType:    taxes.TaxTypeSmallBusiness.String(),
		YearIncome: "0",
	}, svc.calcReqs[0])
}

func TestBotCalculateKeepsFirstIncomeMonth(t *testing.T) {
	h := newBotHarness(t, &fakeService{calcResp: &service.CalculateResponse{}})

	start := func() {
		h.text("/calculate")
		h.press("cb:calc_tax:" + taxes.TaxTypeSmallBusiness.String())
		h.text("0")
		h.press("cb:calc_year:2024")
	}

	var firstTypeMessage int

	addFirstIncome := func() {
		start()
		require.Len(t, h.api.lastButtons(t), 12, "the first income can be in any month")
		h.press("cb:calc_month:March")
		h.press("cb:calc_day:15")
		h.text("100")
		h.press("cb:calc_cur:" + currencies.USD)
		firstTypeMessage = h.api.lastKeyboardMessageID()
		h.press("cb:calc_inc_type:" + taxes.IncomeTypeNonCash.String())
		h.press("cb:calc_more:yes")
		h.press("cb:calc_year:2024")
	}

	addFirstIncome()
	assert.Equal(t, []string{"cb:calc_month:March"}, h.api.lastButtons(t), "later incomes use the first income's month")

	h.press("cb:calc_month:April")
	assert.Contains(t, h.api.lastText(t), "Unexpected input")

	h.press("cb:calc_month:March")
	h.press("cb:calc_day:15")
	h.text("50")
	h.press("cb:calc_cur:" + currencies.USD)
	require.Equal(t, calcStepIncomeType, h.store.get(testUserID).calcStep)

	h.pressOn(firstTypeMessage, "cb:calc_inc_type:"+taxes.IncomeTypeCash.String())
	assert.Contains(t, h.api.lastText(t), "Unexpected input", "a type button of the first income's message is rejected")
	assert.Equal(t, calcStepIncomeType, h.store.get(testUserID).calcStep)

	h.text("/cancel")
	start()
	assert.Len(t, h.api.lastButtons(t), 12, "/cancel at the type step clears the fixed month")

	h.text("/cancel")
	addFirstIncome()
	h.press("cb:calc_month:March")
	h.press("cb:calc_day:15")
	h.text("50")
	h.press("cb:calc_cur:" + currencies.USD)
	h.press("cb:calc_inc_type:" + taxes.IncomeTypeCash.String())
	h.press("cb:calc_more:no")
	h.press("cb:calc_ok:no")
	h.press("cb:calc_year:2024")
	assert.Len(t, h.api.lastButtons(t), 12, "restarting income entry offers all months again")
}

func TestBotCalculateFailedAndPanicked(t *testing.T) {
	run := func(svc *fakeService) *botHarness {
		h := newBotHarness(t, svc)

		h.text("/calculate")
		h.press("cb:calc_tax:" + taxes.TaxTypeEmployment.String())
		h.text("0")
		h.press("cb:calc_year:2024")
		h.press("cb:calc_month:March")
		h.press("cb:calc_day:15")
		h.text("100")
		h.press("cb:calc_cur:" + currencies.GEL)
		h.press("cb:calc_more:no")
		h.press("cb:calc_ok:yes")

		return h
	}

	failed := run(&fakeService{err: errors.New("rates are not available")})
	assert.Contains(t, failed.api.lastText(t), "❌ Calculation error: rates are not available")

	panicked := run(&fakeService{panicMsg: "boom"})
	assert.Equal(t, msgInternalError, panicked.api.lastText(t), "a panic is answered and the bot keeps running")

	panicked.text("/start")
	assert.Contains(t, panicked.api.lastText(t), "Welcome", "the bot still handles updates after a panic")
}

func TestBotIgnoresMessageWithoutSender(t *testing.T) {
	h := newBotHarness(t, &fakeService{})

	h.send(telego.Update{Message: &telego.Message{
		MessageID: 1,
		Chat:      telego.Chat{ID: testChatID, Type: telego.ChatTypeChannel},
		Text:      "/start",
	}})

	assert.Empty(t, h.api.callsOf("sendMessage"))
}
