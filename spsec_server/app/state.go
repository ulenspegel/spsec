package app

import (
    "fmt"
    "time"

    "spsec/config"
)

func (a *App) stateToStr(s int) string {
    names := []string{"закрыто", "открыто", "нет сигнала", "ошибка"}
    if s >= 0 && s < len(names) {
        return names[s]
    }
    return "неизвестно"
}

func (a *App) handleHeartbeat(ts int64) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.lastMsgTime = time.Unix(ts, 0)

    if a.srv.LastState == nil {
        return
    }

    state := *a.srv.LastState

    if state != a.lastHeartbeatState {
        a.lastHeartbeatState = state

        if a.lastState == 2 && state != 2 {
            now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
            a.bot.UpdatePanel(fmt.Sprintf(
                "[%s] ✅ Сигнал восстановлен (%s)",
                now.Format("02.01 15:04:05"),
                a.stateToStr(state),
            ))
        }

        a.notifyState(state)
    }
}
