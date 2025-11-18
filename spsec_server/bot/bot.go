package bot

import (
    "fmt"
    "log"
    "strconv"

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

        // INLINE CALLBACK
        if u.CallbackQuery != nil {
            b.handleCallback(u.CallbackQuery)

            // обновляем панель
            if b.OnStatus != nil {
                b.UpdatePanel(b.OnStatus())
            }
            continue
        }

        if u.Message == nil {
            continue
        }

        b.LastChatID = u.Message.Chat.ID

        // удаляем любое не-командное сообщение
        if !u.Message.IsCommand() {
            _, _ = b.BotAPI.Request(tgbotapi.NewDeleteMessage(b.LastChatID, u.Message.MessageID))
        }

        if u.Message.IsCommand() {
            b.handleCommand(u.Message)
        } else {
            b.handleText(u.Message)
        }

        // обновляем панель
        if b.OnStatus != nil {
            b.UpdatePanel(b.OnStatus())
        }
    }
}

// ============================================================================
//  DYNAMIC PANEL
// ============================================================================

func (b *Bot) UpdatePanel(text string) {
    if b.LastChatID == 0 {
        return // ещё не получали сообщений → нельзя создать панель
    }

    kb := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Обновить", "panel_refresh"),
        ),
    )

    // если панели нет — создаём
    if b.panelMsgID == 0 {
        msg := tgbotapi.NewMessage(b.LastChatID, text)
        msg.ReplyMarkup = kb

        sent, err := b.BotAPI.Send(msg)
        if err != nil {
            log.Println("cannot create panel:", err)
            return
        }

        b.panelMsgID = sent.MessageID
        return
    }

    // обновляем существующую панель
    edit := tgbotapi.NewEditMessageText(b.LastChatID, b.panelMsgID, text)
    edit.ReplyMarkup = &kb

    _, err := b.BotAPI.Send(edit)
    if err != nil {
        log.Println("panel edit failed, recreating:", err)

        // удаляем ID и создаём панель заново
        b.panelMsgID = 0
        b.UpdatePanel(text)
    }
}

// ============================================================================
//  COMMANDS
// ============================================================================

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
    switch msg.Command() {

    case "status":
        if b.OnStatus != nil {
            b.Send(b.OnStatus())
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

// ============================================================================
//  CALLBACKS (INCLUDING PANEL)
// ============================================================================

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
    data := cb.Data

    // panel refresh
    if data == "panel_refresh" {
        if b.OnStatus != nil {
            b.UpdatePanel(b.OnStatus())
        }
        b.answerCallback(cb, "Обновлено")
        return
    }

    // select mode
    if len(data) > 5 && data[:5] == "mode_" {
        mid, _ := strconv.Atoi(data[5:])
        if b.OnModeChange != nil {
            b.OnModeChange(mid)
        }
        b.answerCallback(cb, fmt.Sprintf("Режим переключён на %d", mid))
        return
    }

    b.answerCallback(cb, "Неизвестная команда")
}

func (b *Bot) answerCallback(cb *tgbotapi.CallbackQuery, text string) {
    ans := tgbotapi.NewCallback(cb.ID, text)
    b.BotAPI.Send(ans)
}

// ============================================================================
//  SETUP TIME FSM
// ============================================================================

func (b *Bot) startSetupTime() {
    b.awaitingStart = true
    b.awaitingEnd = false
    b.tmpStart = 0

    b.Send("Введите время старта (0-23):")
}

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

func (b *Bot) handleStartInput(txt string) {
    v, err := strconv.Atoi(txt)
    if err != nil || v < 0 || v > 23 {
        b.Send("Ошибка: нужно число от 0 до 23. /setupTime чтобы начать заново.")
        b.resetSetup()
        return
    }

    b.tmpStart = v
    b.awaitingStart = false
    b.awaitingEnd = true

    b.Send("Введите время окончания (0-23):")
}

func (b *Bot) handleEndInput(txt string) {
    v, err := strconv.Atoi(txt)
    if err != nil || v < 0 || v > 23 {
        b.Send("Ошибка: нужно число от 0 до 23. /setupTime чтобы начать заново.")
        b.resetSetup()
        return
    }

    start := b.tmpStart
    end := v

    b.resetSetup()

    if b.OnScheduleChange != nil {
        b.OnScheduleChange(start, end)
    }

    b.Send(fmt.Sprintf("Расписание изменено: %d → %d", start, end))
}

func (b *Bot) resetSetup() {
    b.awaitingStart = false
    b.awaitingEnd = false
    b.tmpStart = 0
}

// ============================================================================
//  SEND
// ============================================================================

func (b *Bot) Send(text string) {
    if b.LastChatID == 0 {
        return
    }
    msg := tgbotapi.NewMessage(b.LastChatID, text)
    b.BotAPI.Send(msg)
}
