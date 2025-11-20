package bot

import (
    "fmt"
    "strconv"
)

func (b *Bot) startSetupTime() {
    b.awaitingStart = true
    b.awaitingEnd = false
    b.tmpStart = 0

    b.Send("Введите время старта (0-23):")
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
