package app

import (
    "fmt"
    "time"

    "spsec/config"
)

func (a *App) sendState(state int) {
    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] Новое состояние: %s",
        now.Format("02.01 15:04:05"),
        a.stateToStr(state),
    ))
}

func (a *App) notifyState(state int) {
    if state == a.lastState {
        return
    }
    a.lastState = state

    // Записываем в лог
    if err := a.log.Add(state); err != nil {
        fmt.Println("Logger.Add error:", err)
    }

    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)

    // Обновляем панель
    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] %s",
        now.Format("02.01 15:04:05"),
        a.stateToStr(state),
    ))

    // Если режим Alarm, отправляем отдельное сообщение
    if a.currentModeID == 1 {
        a.bot.Send(fmt.Sprintf("⚠️ Состояние изменилось: %s", a.stateToStr(state)))
    }
}


