package app

import (
    // "fmt"
    "time"

    // "spsec/config"
)

func (a *App) watchdog() {
    for {
        time.Sleep(2 * time.Second)

        a.mu.Lock()
        last := a.lastMsgTime
        a.mu.Unlock()

        if time.Since(last) > 5*time.Second {
            a.triggerTimeout()
        }
    }
}


func (a *App) triggerTimeout() {
    a.mu.Lock()
    defer a.mu.Unlock()

    // Если уже был оффлайн — повторно не тревожим
    if a.lastHeartbeatState == 2 {
        return
    }

    a.lastHeartbeatState = 2

    // Панель обновляем через notifyState
    a.notifyState(a.lastState)

    // ОТДЕЛЬНЫМ сообщением — тревога!!!
    a.bot.Send("⚠️ Потеря сигнала")
}


