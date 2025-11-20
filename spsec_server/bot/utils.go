package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) Send(text string) {
    if b.LastChatID == 0 {
        return
    }
    msg := tgbotapi.NewMessage(b.LastChatID, text)
    b.BotAPI.Send(msg)
}
