package app

import (
    "log"
    "spsec/mode"
)

func (a *App) applyMode() {
    switch a.currentModeID {
    case 0:
        a.mode = mode.NewSilent()
    case 1:
        a.mode = mode.NewAlarm(a.sendState)
    case 2:
        a.mode = mode.NewSchedule(a.scheduleStart, a.scheduleEnd, a.sendState)
    }
    log.Println("Режим установлен:", a.mode.Name())
}
