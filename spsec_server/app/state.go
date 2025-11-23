package app

import (
    "fmt"
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
    if a.srv.LastState == nil {
        return
    }

    state := *a.srv.LastState

    restored := a.lastHeartbeatState == 2 && state != 2
    if state != a.lastHeartbeatState || restored {
        a.lastHeartbeatState = state
    
        if restored {
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


