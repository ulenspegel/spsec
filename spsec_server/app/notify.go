package app

import (
    "fmt"
    "time"

    "spsec/config"
)

func (a *App) sendState(state int) {
    // now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
    // a.bot.UpdatePanel(fmt.Sprintf(
    //     "[%s] Новое состояние: %s",
    //     now.Format("02.01 15:04:05"),
    //     a.stateToStr(state),
    // ))
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

    // Сеть: доступна или нет
    networkStatus := "✖️ недоступна"
    if a.lastHeartbeatState != 2 {
        networkStatus = "✔️ доступна"
    }

    // Время последнего сообщения
    lastTime := a.lastMsgTime.UTC().Add(time.Duration(config.GMT) * time.Hour).Format("02.01 15:04:05")

    // Режим по ID через мэппинг
    modeName := getModeName(a.currentModeID)

    // Статус
    status := a.stateToStr(state)

    // Формируем текст панели
    panelText := fmt.Sprintf(
        "Сеть: %s\nВремя: %s\nРежим: %s\nСтатус: %s",
        networkStatus,
        lastTime,
        modeName,
        status,
    )

    // Обновляем панель
    a.bot.UpdatePanel(panelText)

    // Если режим Alarm (ID = 1), отправляем отдельное сообщение
    if a.currentModeID == 1 {
        a.bot.Send(status)
    }
}

// Мэппинг ID режима в название
func getModeName(id int) string {
    names := map[int]string{
        0: "стандартный",
        1: "🚨 активный",
        2: "режим 2",
        // добавь сюда все твои режимы
    }
    if n, ok := names[id]; ok {
        return n
    }
    return "–"
}




