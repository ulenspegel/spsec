package app

import (
    "fmt"
    "log"
    "time"
    "spsec/config"
)

func (a *App) stateToStr(s int) string {
    
    names := []string{"🚪закрыта ✅", "🚪открыта 🛑", "⚠️ нет сети", "📶 сеть восстановлена"}
    if s >= 0 && s < len(names) {
        return names[s]
    }
    return "неизвестно"
}

func (a *App) handleHeartbeat(ts int64) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.lastMsgTime = time.Unix(ts, 0)

    // Если были в состоянии "нет сигнала" (2), а теперь получили heartbeat
    if a.lastState == 2 {
        // Но состояние двери известно через lastHeartbeatState (0 или 1)
        restoredState := a.lastHeartbeatState

        // Обновляем внутреннее состояние
        a.notifyState(restoredState)

        now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
        a.bot.UpdatePanel(fmt.Sprintf(
            "[%s] ✅ Сигнал восстановлен (%s)",
            now.Format("02.01 15:04:05"),
            a.stateToStr(restoredState),
        ))
    }
}

func (a *App) handleNewState(state int, ts int64) {
    a.mu.Lock()
    defer a.mu.Unlock()

    // Update heartbeat tracking
    a.lastMsgTime = time.Unix(ts, 0)
    a.lastHeartbeatState = state

    // Notify about state change
    a.notifyState(state)

    log.Printf("State change handled: %d -> %d", a.lastState, state)
}


