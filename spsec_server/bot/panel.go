package bot

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
func (b *Bot) UpdatePanel(text string) {
    if b.LastChatID == 0 {
        return // ещё не получали сообщений
    }

    kb := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Обновить", "panel_refresh"),
        ),
    )

    // если панель уже есть — удаляем старую
    if b.panelMsgID != 0 {
        _, _ = b.BotAPI.Request(tgbotapi.NewDeleteMessage(b.LastChatID, b.panelMsgID))
        b.panelMsgID = 0
    }

    // создаём новую панель
    msg := tgbotapi.NewMessage(b.LastChatID, text)
    msg.ReplyMarkup = kb
    sent, err := b.BotAPI.Send(msg)
    if err != nil {
        log.Println("cannot create panel:", err)
        return
    }

    b.panelMsgID = sent.MessageID
}
