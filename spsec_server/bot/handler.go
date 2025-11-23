package bot

import (
    "fmt"
    "strconv"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ---------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
    switch msg.Command() {

    case "status":
        if b.OnStatus != nil {
            b.OnStatus();
            //b.Send(b.OnStatus())
        }

    case "mode":
        b.sendModeMenu()

    case "setupTime":
        b.startSetupTime()

    default:
        b.Send("Неизвестная команда")
    }
}

func (b *Bot) sendModeMenu() {
    kb := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("silent", "mode_0"),
            tgbotapi.NewInlineKeyboardButtonData("alarm", "mode_1"),
            tgbotapi.NewInlineKeyboardButtonData("schedule", "mode_2"),
        ),
    )

    msg := tgbotapi.NewMessage(b.LastChatID, "Выберите режим:")
    msg.ReplyMarkup = kb
    b.BotAPI.Send(msg)
}

// ---------------------------------------------------------------------
// Callbacks
// ---------------------------------------------------------------------

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
    data := cb.Data

    switch {
    case data == "panel_refresh":
        if b.OnStatus != nil {
            b.UpdatePanel(b.OnStatus())
        }
        b.answerCallback(cb, "Обновлено")
        return // <- возвращаемся, чтобы не дублировать UpdatePanel

    case len(data) > 5 && data[:5] == "mode_":
        mid, _ := strconv.Atoi(data[5:])
        if b.OnModeChange != nil {
            b.OnModeChange(mid)
        }
        b.answerCallback(cb, fmt.Sprintf("Режим переключён на %d", mid))

    default:
        b.answerCallback(cb, "Неизвестная команда")
    }
}

func (b *Bot) answerCallback(cb *tgbotapi.CallbackQuery, text string) {
    ans := tgbotapi.NewCallback(cb.ID, text)
    b.BotAPI.Send(ans)
}

// ---------------------------------------------------------------------
// Plain text
// ---------------------------------------------------------------------

func (b *Bot) handleText(msg *tgbotapi.Message) {
    if b.awaitingStart {
        b.handleStartInput(msg.Text)
        return
    }

    if b.awaitingEnd {
        b.handleEndInput(msg.Text)
        return
    }
}
