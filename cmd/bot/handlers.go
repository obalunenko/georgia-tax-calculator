package main

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	log "github.com/obalunenko/logger"

	"github.com/obalunenko/georgia-tax-calculator/internal/service"
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

		kb, err := taxTypeKeyboard()
		if err != nil {
			return fmt.Errorf("build tax type keyboard: %w", err)
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: msg.Chat.ID},
			Text:        "💼 Select your tax type:",
			ReplyMarkup: &kb,
		})

		return err
	}
}

// handleConvert handles the /convert command.
func handleConvert(store *sessionStore) telegohandler.MessageHandler {
	return func(ctx *telegohandler.Context, msg telego.Message) error {
		sess := store.get(msg.From.ID)
		sess.flow = flowConvert
		sess.convertStep = convertStepYear
		sess.convertReq = service.ConvertRequest{}

		kb := yearKeyboard()

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: msg.Chat.ID},
			Text:        "📅 Select the year of conversion:",
			ReplyMarkup: &kb,
		})

		return err
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
		if err := validateMoney(text); err != nil {
			_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.Chat.ID},
				Text:   fmt.Sprintf("❌ Invalid amount: %v\n\nPlease enter a valid number (e.g. 1500.00):", err),
			})

			return sendErr
		}

		sess.calcReq.YearIncome = text
		sess.calcStep = calcStepYear

		kb := yearKeyboard()

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: msg.Chat.ID},
			Text:        fmt.Sprintf("✅ Year income set to: %s GEL\n\n📅 Select the year of income:", text),
			ReplyMarkup: &kb,
		})

		return err
	case calcStepAmount:
		text := strings.TrimSpace(msg.Text)
		if err := validateMoney(text); err != nil {
			_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.Chat.ID},
				Text:   fmt.Sprintf("❌ Invalid amount: %v\n\nPlease enter a valid number (e.g. 1500.00):", err),
			})

			return sendErr
		}

		sess.currentInc.Amount = text
		sess.calcStep = calcStepCurrency

		kb := currencyKeyboard()

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: msg.Chat.ID},
			Text:        fmt.Sprintf("✅ Amount set to: %s\n\n💱 Select the currency of income:", text),
			ReplyMarkup: &kb,
		})

		return err
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
		_, sendErr := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: msg.Chat.ID},
			Text:   fmt.Sprintf("❌ Invalid amount: %v\n\nPlease enter a valid number (e.g. 1500.00):", err),
		})

		return sendErr
	}

	sess.convertReq.Amount = text
	sess.convertStep = convertStepCurrencyFrom

	kb := currencyKeyboard()

	_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID:      telego.ChatID{ID: msg.Chat.ID},
		Text:        fmt.Sprintf("✅ Amount set to: %s\n\n💱 Select the source currency:", text),
		ReplyMarkup: &kb,
	})

	return err
}

// handleCallback handles all inline keyboard callback queries.
func handleCallback(store *sessionStore, svc service.Service) telegohandler.CallbackQueryHandler {
	return func(ctx *telegohandler.Context, query telego.CallbackQuery) error {
		// Always answer the callback query to remove the loading indicator.
		if err := ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		}); err != nil {
			log.WithError(ctx.Context(), err).Error("answer callback query")
		}

		if query.Message == nil {
			return nil
		}

		if query.From.ID == 0 {
			return nil
		}

		data := strings.TrimPrefix(query.Data, callbackPrefix)
		chatID := query.Message.GetChat().ID
		userID := query.From.ID

		sess := store.get(userID)

		switch sess.flow {
		case flowCalculate:
			return handleCalcCallback(ctx, chatID, data, sess, svc)
		case flowConvert:
			return handleConvertCallback(ctx, chatID, data, sess, svc)
		default:
			_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   "Please use /calculate or /convert to start.",
			})

			return err
		}
	}
}

func handleCalcCallback(
	ctx *telegohandler.Context,
	chatID int64,
	data string,
	sess *session,
	svc service.Service,
) error {
	switch sess.calcStep {
	case calcStepTaxType:
		taxType := taxTypeFromItem(data)
		sess.calcReq.TaxType = taxType
		sess.calcStep = calcStepYearIncome

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text: fmt.Sprintf(
				"✅ Tax type: %s\n\n💰 Enter your income from the beginning of the calendar year in GEL:\n(e.g. 0.00 if this is your first income)",
				taxType,
			),
		})

		return err

	case calcStepYear:
		sess.currentInc.Year = data
		sess.calcStep = calcStepMonth

		kb, err := monthKeyboard(data)
		if err != nil {
			return fmt.Errorf("build month keyboard: %w", err)
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Year: %s\n\n📅 Select the month of income:", data),
			ReplyMarkup: &kb,
		})

		return err

	case calcStepMonth:
		sess.currentInc.Month = data
		sess.calcStep = calcStepDay

		kb, err := dayKeyboard(sess.currentInc.Year, data)
		if err != nil {
			return fmt.Errorf("build day keyboard: %w", err)
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Month: %s\n\n📅 Select the day of income:", data),
			ReplyMarkup: &kb,
		})

		return err

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
		sess.incomes = append(sess.incomes, sess.currentInc)
		sess.currentInc = service.Income{}
		sess.calcStep = calcStepAddMore

		incomeList := formatIncomeList(sess.incomes)

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Currency: %s\n\n%s\n➕ Add another income entry?", data, incomeList),
			ReplyMarkup: buildConfirmKeyboardPtr(),
		})

		return err

	case calcStepAddMore:
		if data == confirmYes {
			sess.calcStep = calcStepYear

			kb := yearKeyboard()

			_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID:      telego.ChatID{ID: chatID},
				Text:        "📅 Select the year of income:",
				ReplyMarkup: &kb,
			})

			return err
		}

		// Show summary and ask for confirmation.
		sess.calcStep = calcStepConfirm

		summary := formatCalcSummary(sess)

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("📋 Review your inputs:\n\n%s\n\nAre your answers correct?", summary),
			ReplyMarkup: buildConfirmKeyboardPtr(),
		})

		return err

	case calcStepConfirm:
		if data == confirmNo {
			// Restart the whole flow.
			sess.calcStep = calcStepTaxType
			sess.calcReq = service.CalculateRequest{}
			sess.incomes = nil
			sess.currentInc = service.Income{}

			kb, err := taxTypeKeyboard()
			if err != nil {
				return fmt.Errorf("build tax type keyboard: %w", err)
			}

			_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID:      telego.ChatID{ID: chatID},
				Text:        "🔄 Restarting...\n\n💼 Select your tax type:",
				ReplyMarkup: &kb,
			})

			return err
		}

		// Calculate taxes.
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

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   formatCalcResult(resp),
		})

		return err

	default:
		return sendUnexpectedInput(ctx, chatID)
	}
}

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

		kb, err := monthKeyboard(data)
		if err != nil {
			return fmt.Errorf("build month keyboard: %w", err)
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Year: %s\n\n📅 Select the month:", data),
			ReplyMarkup: &kb,
		})

		return err

	case convertStepMonth:
		sess.convertReq.Month = data
		sess.convertStep = convertStepDay

		kb, err := dayKeyboard(sess.convertReq.Year, data)
		if err != nil {
			return fmt.Errorf("build day keyboard: %w", err)
		}

		_, err = sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Month: %s\n\n📅 Select the day:", data),
			ReplyMarkup: &kb,
		})

		return err

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

		kb := currencyKeyboard()

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("✅ Source currency: %s\n\n💱 Select the target currency:", data),
			ReplyMarkup: &kb,
		})

		return err

	case convertStepCurrencyTo:
		sess.convertReq.CurrencyTo = data
		sess.convertStep = convertStepConfirm

		summary := formatConvertSummary(sess.convertReq)

		_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
			ChatID:      telego.ChatID{ID: chatID},
			Text:        fmt.Sprintf("📋 Review your inputs:\n\n%s\n\nAre your answers correct?", summary),
			ReplyMarkup: buildConfirmKeyboardPtr(),
		})

		return err

	case convertStepConfirm:
		if data == confirmNo {
			sess.convertStep = convertStepYear
			sess.convertReq = service.ConvertRequest{}

			kb := yearKeyboard()

			_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
				ChatID:      telego.ChatID{ID: chatID},
				Text:        "🔄 Restarting...\n\n📅 Select the year of conversion:",
				ReplyMarkup: &kb,
			})

			return err
		}

		// Run conversion.
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

	default:
		return sendUnexpectedInput(ctx, chatID)
	}
}

func sendUnexpectedInput(ctx *telegohandler.Context, chatID int64) error {
	_, err := sendMessage(ctx, ctx.Bot(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⚠️ Unexpected input. Please follow the flow or use /cancel to start over.",
	})

	return err
}

// buildConfirmKeyboardPtr returns a pointer to an InlineKeyboardMarkup for yes/no.
func buildConfirmKeyboardPtr() *telego.InlineKeyboardMarkup {
	kb := buildConfirmKeyboard()

	return &kb
}

// formatIncomeList formats the list of captured incomes for display.
func formatIncomeList(incomes []service.Income) string {
	if len(incomes) == 0 {
		return "Captured incomes: none yet"
	}

	var b strings.Builder

	b.WriteString("Captured incomes:\n")

	for i, inc := range incomes {
		b.WriteString(fmt.Sprintf("  %d) %s-%s-%s — %s %s\n",
			i+1, inc.Year, inc.Month, inc.Day, inc.Amount, inc.Currency))
	}

	return b.String()
}

// formatCalcSummary formats the calculation request summary.
func formatCalcSummary(sess *session) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Tax type: %s\n", sess.calcReq.TaxType))
	b.WriteString(fmt.Sprintf("Year income (GEL): %s\n", sess.calcReq.YearIncome))
	b.WriteString("\nIncomes:\n")

	for i, inc := range sess.incomes {
		b.WriteString(fmt.Sprintf("  %d) %s-%s-%s — %s %s\n",
			i+1, inc.Year, inc.Month, inc.Day, inc.Amount, inc.Currency))
	}

	return b.String()
}

// formatCalcResult formats the calculation response.
func formatCalcResult(resp *service.CalculateResponse) string {
	var b strings.Builder

	b.WriteString("🧾 Tax Calculation Result\n\n")
	b.WriteString(fmt.Sprintf("Tax Rate: %s\n", resp.TaxRate.String()))
	b.WriteString(fmt.Sprintf("Year Income: %s\n", resp.YearIncome.String()))

	if len(resp.Incomes) > 0 {
		b.WriteString("\nIncomes converted to GEL:\n")

		for i, inc := range resp.Incomes {
			b.WriteString(fmt.Sprintf("  %d) %s → %s (rate: %s)\n",
				i+1,
				inc.Amount.String(),
				inc.Converted.String(),
				inc.Rate.String(),
			))
		}
	}

	b.WriteString(fmt.Sprintf("\nTotal Income (GEL): %s\n", resp.TotalIncomeConverted.String()))
	b.WriteString(fmt.Sprintf("Taxes to Pay: %s", resp.Tax.String()))

	return b.String()
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

// validateMoney validates that s is a valid money amount.
func validateMoney(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("value is required")
	}

	if _, err := parseMoneyAmount(s); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	return nil
}

// parseMoneyAmount parses a money amount from string.
func parseMoneyAmount(s string) (float64, error) {
	var f float64

	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0, fmt.Errorf("parse amount: %w", err)
	}

	return f, nil
}
