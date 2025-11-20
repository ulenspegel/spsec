package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	BotAPI     *tgbotapi.BotAPI
	LastChatID int64

	// external callbacks
	OnStatus         func() string
	OnModeChange     func(int)
	OnScheduleChange func(int, int)

	// internal FSM
	awaitingStart bool
	awaitingEnd   bool
	tmpStart      int

	// dynamic panel
	panelMsgID int
}

func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{BotAPI: api}, nil
}

func (b *Bot) Start() {
	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	updates := b.BotAPI.GetUpdatesChan(upd)

	for u := range updates {

		if u.CallbackQuery != nil {
			b.handleCallback(u.CallbackQuery)
			continue // больше не нужно UpdatePanel здесь
		}

		if u.Message == nil {
			continue
		}

		b.LastChatID = u.Message.Chat.ID

		if !u.Message.IsCommand() {
			_, _ = b.BotAPI.Request(
				tgbotapi.NewDeleteMessage(b.LastChatID, u.Message.MessageID),
			)
		}

		if u.Message.IsCommand() {
			b.handleCommand(u.Message)
		} else {
			b.handleText(u.Message)
		}

		if b.OnStatus != nil {
			b.UpdatePanel(b.OnStatus())
		}
	}
}
