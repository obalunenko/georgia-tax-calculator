package main

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	log "github.com/obalunenko/logger"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
)

const (
	cmdStart     = "start"
	cmdCalculate = "calculate"
	cmdConvert   = "convert"
	cmdCancel    = "cancel"
	cmdHelp      = "help"
)

// handleStart handles the /start command.
func handleStart(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		store.reset(msg.From.ID)

		text := "👋 Welcome to Georgia Tax Calculator Bot!\n\n" +
			"I can help you calculate taxes and convert currencies according to official NBG rates.\n\n" +
			"Available commands:\n" +
			"• /calculate — Calculate taxes\n" +
			"• /convert — Convert currency\n" +
			"• /cancel — Cancel current operation\n" +
			"• /help — Show this help message"

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: msg.Chat.ID},
			Text:   text,
		})

		return err
	}
}

// handleHelp handles the /help command.
func handleHelp(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		store.reset(msg.From.ID)

		text := "🧮 Georgia Tax Calculator Bot\n\n" +
			"Commands:\n" +
			"• /calculate — Start tax calculation flow\n" +
			"  Calculates your taxes based on income, currency, and tax type\n\n" +
			"• /convert — Start currency conversion flow\n" +
			"  Converts an amount using the official NBG exchange rate for a given date\n\n" +
			"• /cancel — Cancel current operation and reset\n\n" +
			"• /help — Show this help message"

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: msg.Chat.ID},
			Text:   text,
		})

		return err
	}
}

// handleCancel handles the /cancel command.
func handleCancel(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		store.reset(msg.From.ID)

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: msg.Chat.ID},
			Text:   "✅ Operation cancelled. Use /calculate or /convert to start again.",
		})

		return err
	}
}

// handleCalculate handles the /calculate command.
func handleCalculate(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		sess := store.get(msg.From.ID)
		sess.flow = flowCalculate
		sess.calcStep = calcStepTaxType
		sess.calcReq = service.CalculateRequest{}
		sess.incomes = nil
		sess.currentInc = service.Income{}

		return sendStep(ctx, msg.Chat.ID, sess,
			"📄 Declarations: "+service.DeclarationsURL+"\n\n💼 Select your tax type:")
	}
}

// handleConvert handles the /convert command.
func handleConvert(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		sess := store.get(msg.From.ID)
		sess.flow = flowConvert
		sess.convertStep = convertStepYear
		sess.convertReq = service.ConvertRequest{}

		return sendStep(ctx, msg.Chat.ID, sess, "📅 Select the year of conversion:")
	}
}

// handleTextInput handles free-text messages (for amount input).
func handleTextInput(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		sess := store.get(msg.From.ID)

		switch sess.flow {
		case flowCalculate:
			return handleCalcTextInput(ctx, msg, sess)
		case flowConvert:
			return handleConvertTextInput(ctx, msg, sess)
		default:
			_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.Chat.ID},
				Text:   "Please use /calculate or /convert to start.",
			})

			return err
		}
	}
}

func handleCalcTextInput(
	ctx *telegohandler.Context,
	msg telego.Message,
	sess *session,
) error {
	switch sess.calcStep {
	case calcStepYearIncome:
		text := strings.TrimSpace(msg.Text)
		if err := validateYearIncome(text); err != nil {
			return sendInvalidAmount(ctx, msg.Chat.ID, err)
		}

		sess.calcReq.YearIncome = text
		sess.calcStep = calcStepYear

		return sendStep(ctx, msg.Chat.ID, sess,
			fmt.Sprintf("✅ Year income set to: %s GEL\n\n📅 Select the year of income:", text))
	case calcStepAmount:
		text := strings.TrimSpace(msg.Text)
		if err := validateMoney(text); err != nil {
			return sendInvalidAmount(ctx, msg.Chat.ID, err)
		}

		sess.currentInc.Amount = text
		sess.calcStep = calcStepCurrency

		return sendStep(ctx, msg.Chat.ID, sess,
			fmt.Sprintf("✅ Amount set to: %s\n\n💱 Select the currency of income:", text))
	default:
		return sendUnexpectedInput(ctx, msg.Chat.ID)
	}
}

func handleConvertTextInput(
	ctx *telegohandler.Context,
	msg telego.Message,
	sess *session,
) error {
	if sess.convertStep != convertStepAmount {
		return sendUnexpectedInput(ctx, msg.Chat.ID)
	}

	text := strings.TrimSpace(msg.Text)
	if err := validateMoney(text); err != nil {
		return sendInvalidAmount(ctx, msg.Chat.ID, err)
	}

	sess.convertReq.Amount = text
	sess.convertStep = convertStepCurrencyFrom

	return sendStep(ctx, msg.Chat.ID, sess,
		fmt.Sprintf("✅ Amount set to: %s\n\n💱 Select the source currency:", text))
}

// answerCallback answers the callback query before handling it, so the loading indicator
// is removed even if the handler waits for another update of the same user or fails.
func answerCallback(h telegohandler.CallbackQueryHandler) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		if err := ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		}); err != nil {
			log.WithError(ctx.Context(), err).Error("answer callback query")
		}

		return h(ctx, query)
	}
}

// handleCallback handles all inline keyboard callback queries.
func handleCallback(store *sessionStore, svc service.Service) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		if query.Message == nil {
			return nil
		}

		if query.From.ID == 0 {
			return nil
		}

		chatID := query.Message.GetChat().ID
		userID := query.From.ID

		sess := store.get(userID)

		if sess.flow == flowNone {
			_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   "Please use /calculate or /convert to start.",
			})

			return err
		}

		// Only the latest keyboard is accepted, so a button of an earlier income's message
		// cannot answer the same step of a later income.
		if query.Message.GetMessageID() != sess.keyboardMsgID {
			return sendUnexpectedInput(ctx, chatID)
		}

		data, ok, err := callbackValue(sess, query.Data)
		if err != nil {
			return fmt.Errorf("check callback data: %w", err)
		}

		if !ok {
			return sendUnexpectedInput(ctx, chatID)
		}

		switch sess.flow {
		case flowCalculate:
			return handleCalcCallback(ctx, chatID, data, sess, svc)
		case flowConvert:
			return handleConvertCallback(ctx, chatID, data, sess, svc)
		case flowNone:
		}

		return sendUnexpectedInput(ctx, chatID)
	}
}

// handleCalcCallback applies data, already checked by callbackValue, to the current calculation step.
func handleCalcCallback(
	ctx *telegohandler.Context,
	chatID int64,
	data string,
	sess *session,
	svc service.Service,
) error {
	switch sess.calcStep {
	case calcStepTaxType:
		sess.calcReq.TaxType = data
		sess.calcStep = calcStepYearIncome

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   fmt.Sprintf("✅ Tax type: %s\n\n%s", data, yearIncomeQuestion(data)),
		})

		return err

	case calcStepYear:
		sess.currentInc.Year = data
		sess.calcStep = calcStepMonth

		return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Year: %s\n\n📅 Select the month of income:", data))

	case calcStepMonth:
		sess.currentInc.Month = data
		sess.calcStep = calcStepDay

		return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Month: %s\n\n📅 Select the day of income:", data))

	case calcStepDay:
		sess.currentInc.Day = data
		sess.calcStep = calcStepAmount

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text: fmt.Sprintf(
				"✅ Date: %s-%s-%s\n\n💵 Enter the income amount:\n(e.g. 1500.00)",
				sess.currentInc.Year, sess.currentInc.Month, data,
			),
		})

		return err

	case calcStepCurrency:
		sess.currentInc.Currency = data

		if isSmallBusiness(sess.calcReq.TaxType) {
			sess.calcStep = calcStepIncomeType

			return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Currency: %s\n\n%s", data, incomeTypeQuestion()))
		}

		return addIncome(ctx, chatID, sess, "✅ Currency: "+data)

	case calcStepIncomeType:
		sess.currentInc.Type = data

		return addIncome(ctx, chatID, sess, "✅ Income type: "+taxes.IncomeTypeLabel(data))

	case calcStepAddMore:
		if data == confirmYes {
			sess.calcStep = calcStepYear

			return sendStep(ctx, chatID, sess, "📅 Select the year of income:")
		}

		// Show summary and ask for confirmation.
		sess.calcStep = calcStepConfirm

		summary := formatCalcSummary(sess)

		return sendStep(ctx, chatID, sess, fmt.Sprintf("📋 Review your inputs:\n\n%s\n\nAre your answers correct?", summary))

	case calcStepConfirm:
		if data == confirmNo {
			sess.restartIncomes()

			return sendStep(ctx, chatID, sess, "🔄 Restarting income entry...\n\n📅 Select the year of income:")
		}

		return runCalculation(ctx, chatID, sess, svc)

	default:
		return sendUnexpectedInput(ctx, chatID)
	}
}

// yearIncomeQuestion asks for the income since the start of the year. The Small Business categories enter
// field (15) of the previous month's declaration, so this month's incomes are not counted twice.
func yearIncomeQuestion(taxType string) string {
	if isSmallBusiness(taxType) {
		return "💰 Enter field (15) of your previous month's declaration: income since the start of the year in GEL:\n" +
			"(0.00 for your first declaration of the year)"
	}

	return "💰 Enter your income from the beginning of the calendar year in GEL:\n(e.g. 0.00 if this is your first income)"
}

// addIncome captures the current income and asks whether to add another one.
func addIncome(ctx *telegohandler.Context, chatID int64, sess *session, confirmation string) error {
	sess.incomes = append(sess.incomes, sess.currentInc)
	sess.currentInc = service.Income{}
	sess.calcStep = calcStepAddMore

	return sendStep(ctx, chatID, sess,
		fmt.Sprintf("%s\n\n%s\n➕ Add another income entry?", confirmation, formatIncomeList(sess.incomes)))
}

// incomeTypeQuestion asks for the income type and explains each type, as the CLI options do.
func incomeTypeQuestion() string {
	var b strings.Builder

	b.WriteString("🧾 Select the income type (declaration field):")

	for _, t := range taxes.AllIncomeTypes() {
		fmt.Fprintf(&b, "\n• %s — %s", t.Label(), t.Description())
	}

	return b.String()
}

// runCalculation calculates taxes for the confirmed inputs and ends the flow.
func runCalculation(ctx *telegohandler.Context, chatID int64, sess *session, svc service.Service) error {
	sess.calcReq.Income = sess.incomes
	sess.calcStep = calcStepDone
	sess.flow = flowNone

	_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⏳ Calculating taxes...",
	})
	if err != nil {
		return err
	}

	resp, err := svc.Calculate(ctx.Context(), sess.calcReq)
	if err != nil {
		_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   fmt.Sprintf("❌ Calculation error: %v\n\nPlease try again with /calculate", err),
		})

		return sendErr
	}

	text, entities := formatCalcResult(resp)

	_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID:   telego.ChatID{ID: chatID},
		Text:     text,
		Entities: entities,
	})

	return err
}

// handleConvertCallback applies data, already checked by callbackValue, to the current conversion step.
func handleConvertCallback(
	ctx *telegohandler.Context,
	chatID int64,
	data string,
	sess *session,
	svc service.Service,
) error {
	switch sess.convertStep {
	case convertStepYear:
		sess.convertReq.Year = data
		sess.convertStep = convertStepMonth

		return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Year: %s\n\n📅 Select the month:", data))

	case convertStepMonth:
		sess.convertReq.Month = data
		sess.convertStep = convertStepDay

		return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Month: %s\n\n📅 Select the day:", data))

	case convertStepDay:
		sess.convertReq.Day = data
		sess.convertStep = convertStepAmount

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text: fmt.Sprintf(
				"✅ Date: %s-%s-%s\n\n💵 Enter the amount to convert:\n(e.g. 1500.00)",
				sess.convertReq.Year, sess.convertReq.Month, data,
			),
		})

		return err

	case convertStepCurrencyFrom:
		sess.convertReq.CurrencyFrom = data
		sess.convertStep = convertStepCurrencyTo

		return sendStep(ctx, chatID, sess, fmt.Sprintf("✅ Source currency: %s\n\n💱 Select the target currency:", data))

	case convertStepCurrencyTo:
		sess.convertReq.CurrencyTo = data
		sess.convertStep = convertStepConfirm

		summary := formatConvertSummary(sess.convertReq)

		return sendStep(ctx, chatID, sess, fmt.Sprintf("📋 Review your inputs:\n\n%s\n\nAre your answers correct?", summary))

	case convertStepConfirm:
		if data == confirmNo {
			sess.convertStep = convertStepYear
			sess.convertReq = service.ConvertRequest{}

			return sendStep(ctx, chatID, sess, "🔄 Restarting...\n\n📅 Select the year of conversion:")
		}

		return runConversion(ctx, chatID, sess, svc)

	default:
		return sendUnexpectedInput(ctx, chatID)
	}
}

// runConversion converts the confirmed amount and ends the flow.
func runConversion(ctx *telegohandler.Context, chatID int64, sess *session, svc service.Service) error {
	sess.convertStep = convertStepDone
	sess.flow = flowNone

	_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⏳ Converting...",
	})
	if err != nil {
		return err
	}

	resp, err := svc.Convert(ctx.Context(), sess.convertReq)
	if err != nil {
		_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   fmt.Sprintf("❌ Conversion error: %v\n\nPlease try again with /convert", err),
		})

		return sendErr
	}

	_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   formatConvertResult(resp),
	})

	return err
}

// sendStep sends text with the keyboard of the current step.
func sendStep(ctx *telegohandler.Context, chatID int64, sess *session, text string) error {
	kb, err := stepKeyboard(sess)
	if err != nil {
		return fmt.Errorf("build keyboard: %w", err)
	}

	msg, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID:      telego.ChatID{ID: chatID},
		Text:        text,
		ReplyMarkup: kb,
	})
	if err != nil {
		return err
	}

	sess.keyboardMsgID = msg.MessageID

	return nil
}

func sendUnexpectedInput(ctx *telegohandler.Context, chatID int64) error {
	_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⚠️ Unexpected input. Please follow the flow or use /cancel to start over.",
	})

	return err
}

func sendInvalidAmount(ctx *telegohandler.Context, chatID int64, err error) error {
	_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   invalidAmountText(err),
	})

	return sendErr
}

// invalidAmountText formats a validation error of service.ParseAmount or service.ParseYearIncome.
func invalidAmountText(err error) string {
	msg := err.Error()
	if msg != "" {
		msg = strings.ToUpper(msg[:1]) + msg[1:]
	}

	return fmt.Sprintf("❌ %s\n\nPlease enter a valid number (e.g. 1500.00):", msg)
}

// formatIncomeList formats the list of captured incomes for display.
func formatIncomeList(incomes []service.Income) string {
	if len(incomes) == 0 {
		return "Captured incomes: none yet"
	}

	var b strings.Builder

	b.WriteString("Captured incomes:\n")

	for i, inc := range incomes {
		b.WriteString(formatIncomeLine(i+1, inc))
	}

	return b.String()
}

// formatIncomeLine formats one captured income, with its type when it has one.
func formatIncomeLine(n int, inc service.Income) string {
	line := fmt.Sprintf("  %d) %s-%s-%s — %s %s", n, inc.Year, inc.Month, inc.Day, inc.Amount, inc.Currency)

	if label := taxes.IncomeTypeLabel(inc.Type); label != "" {
		line += " — " + label
	}

	return line + "\n"
}

// formatCalcSummary formats the calculation request summary.
func formatCalcSummary(sess *session) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Tax type: %s\n", sess.calcReq.TaxType))
	b.WriteString(fmt.Sprintf("Year income (GEL): %s\n", sess.calcReq.YearIncome))
	b.WriteString("\nIncomes:\n")

	for i, inc := range sess.incomes {
		b.WriteString(formatIncomeLine(i+1, inc))
	}

	return b.String()
}

// formatCalcResult formats the calculation response. Each declaration value gets a code entity,
// so Telegram copies it with a tap; the rest of the text stays plain.
func formatCalcResult(resp *service.CalculateResponse) (string, []telego.MessageEntity) {
	var b strings.Builder

	b.WriteString("🧾 Tax Calculation Result\n\n")
	b.WriteString(fmt.Sprintf("Tax Rate: %s\n", resp.TaxRate.String()))
	b.WriteString(fmt.Sprintf("Year Income: %s\n", resp.YearIncome.String()))

	if len(resp.Incomes) > 0 {
		b.WriteString("\nIncomes converted to GEL:\n")

		for i, inc := range resp.Incomes {
			fmt.Fprintf(&b, "  %d) %s → %s (rate: %s)",
				i+1,
				inc.Amount.String(),
				inc.Converted.String(),
				inc.Rate.String(),
			)

			if inc.Type.Valid() {
				b.WriteString(" — " + inc.Type.Label())
			}

			b.WriteByte('\n')
		}
	}

	b.WriteString(fmt.Sprintf("\nTotal Income (GEL): %s\n", resp.TotalIncomeConverted.String()))
	b.WriteString(fmt.Sprintf("Taxes to Pay: %s", resp.Tax.String()))

	var entities []telego.MessageEntity

	if resp.Declaration != nil {
		b.WriteString("\n\n")
		b.WriteString(service.DeclarationHeader)

		offset := utf16Len(b.String())

		for _, f := range resp.Declaration.Fields {
			prefix, value := "\n"+f.Prefix(), f.Formatted()
			offset += utf16Len(prefix)

			entities = append(entities, telego.MessageEntity{
				Type:   telego.EntityTypeCode,
				Offset: offset,
				Length: utf16Len(value),
			})

			offset += utf16Len(value)

			b.WriteString(prefix)
			b.WriteString(value)
		}
	}

	if resp.LimitWarning != nil {
		b.WriteString("\n⚠️ ")
		b.WriteString(resp.LimitWarning.String())
	}

	return b.String(), entities
}

// formatConvertSummary formats the convert request for display.
func formatConvertSummary(req service.ConvertRequest) string {
	return fmt.Sprintf("Date: %s-%s-%s\nAmount: %s %s\nConvert to: %s",
		req.Year, req.Month, req.Day,
		req.Amount, req.CurrencyFrom,
		req.CurrencyTo,
	)
}

// formatConvertResult formats the conversion response.
func formatConvertResult(resp *service.ConvertResponse) string {
	var b strings.Builder

	b.WriteString("💱 Currency Conversion Result\n\n")
	b.WriteString(fmt.Sprintf("Date: %s\n", resp.Date.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("Amount: %s\n", resp.Amount.String()))
	b.WriteString(fmt.Sprintf("Converted: %s\n", resp.Converted.String()))
	b.WriteString(fmt.Sprintf("Rate: %s", resp.Rate.String()))

	return b.String()
}

// validateMoney validates an income or conversion amount with the rule shared with the CLI and the service.
func validateMoney(s string) error {
	_, err := service.ParseAmount(s)

	return err
}

// validateYearIncome validates the income since the start of the year with the rule shared with the CLI and the service.
func validateYearIncome(s string) error {
	_, err := service.ParseYearIncome(s)

	return err
}
