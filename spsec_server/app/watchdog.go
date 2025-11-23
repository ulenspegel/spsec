package app

import (
    "fmt"
    "time"

    "spsec/config"
)

func (a *App) watchdog() {
    for {
        time.Sleep(time.Second)

        a.mu.Lock()
        since := time.Since(a.lastMsgTime)
        a.mu.Unlock()

        if since > time.Duration(a.timeoutSeconds)*time.Second {
            a.triggerTimeout()
        }
    }
}

func (a *App) triggerTimeout() {
    a.mu.Lock()
    defer a.mu.Unlock()

    if a.lastHeartbeatState == 2 {
        return
    }

    a.lastHeartbeatState = 2
    a.lastState = 2 // если ты используешь lastState где-то отдельно

    a.notifyState(2)

    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] ⚠️ Потеря сигнала",
        now.Format("02.01 15:04:05"),
    ))
}

