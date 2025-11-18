package app

import (
    "fmt"
    "log"
    "sync"
    "time"

    "spsec/bot"
    "spsec/config"
    "spsec/logger"
    "spsec/mode"
    "spsec/serv"
)

type App struct {
    log   *logger.Logger
    srv   *serv.Server
    bot   *bot.Bot
    mode  mode.Mode

    currentModeID int
    scheduleStart int
    scheduleEnd   int
    lastHeartbeatState int


    lastMsgTime    time.Time
    lastState      int
    timeoutSeconds int
    mu             sync.Mutex
}

func New(log *logger.Logger, srv *serv.Server, bot *bot.Bot, initial mode.Mode) *App {
    a := &App{
        log:            log,
        srv:            srv,
        bot:            bot,
        mode:           initial,
        currentModeID:  0,
        scheduleStart:  9,
        scheduleEnd:    22,
        lastMsgTime:    time.Now(),
        timeoutSeconds: 10,
        lastState:      0,
        lastHeartbeatState: -1, // чтобы отслеживать первое сообщение

    }

    a.initCallbacks()
    return a
}

func (a *App) initCallbacks() {

    // ---------- Telegram -----------

    a.bot.OnStatus = func() string {
        entries := a.log.Last(1)
        if len(entries) == 0 {
            return "Лог пуст"
        }
        last := entries[0]

        localTime := last.Time.UTC().Add(time.Duration(config.GMT) * time.Hour)

        return fmt.Sprintf(
            "Последнее событие: %s - %s",
            localTime.Format("02.01 15:04:05"),
            a.stateToStr(last.State),
        )
    }

    a.bot.OnModeChange = func(mid int) {
        a.currentModeID = mid
        a.applyMode()
    }

    a.bot.OnScheduleChange = func(start, end int) {
        a.scheduleStart = start
        a.scheduleEnd = end
        log.Println("Изменено расписание:", start, end)

        if a.currentModeID == 2 {
            a.mode = mode.NewSchedule(start, end, a.sendState)
            log.Println("Schedule перезапущен")
        }
    }

    // ---------- Server -----------
    //

    a.srv.OnNewState = func(state int, ts int64) {
        a.mu.Lock()
        a.lastMsgTime = time.Unix(ts, 0)
        a.mu.Unlock()
    
        // только передаём состояние дальше в mode,
        // а сообщения отправляет notifyState!
        a.mode.OnState(state)
    }
        
}

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

func (a *App) sendState(state int) {
    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] Новое состояние: %s",
        now.Format("02.01 15:04:05"),
        a.stateToStr(state),
    ))
}

func (a *App) notifyState(state int) {
    // предотвращаем дубли
    if state == a.lastState {
        return
    }

    a.lastState = state

    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)

    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] %s",
        now.Format("02.01 15:04:05"),
        a.stateToStr(state),
    ))
}


func (a *App) stateToStr(s int) string {
    names := []string{"закрыто", "открыто", "нет сигнала", "ошибка"}
    if s >= 0 && s < len(names) {
        return names[s]
    }
    return "неизвестно"
}

func (a *App) Run() {
    go a.bot.Start()
    go a.watchdog()

    a.srv.OnHeartbeat = func(ts int64) {
        a.mu.Lock()
        defer a.mu.Unlock()
    
        a.lastMsgTime = time.Unix(ts, 0)
    
        if a.srv.LastState != nil {
            state := *a.srv.LastState
    
            // если состояние изменилось
            if state != a.lastHeartbeatState {
                a.lastHeartbeatState = state
    
                // восстановление после таймаута
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
    }
    

    // восстановление последнего состояния
    if entries, err := a.log.LoadFromDisk(); err == nil && len(entries) > 0 {
        last := entries[len(entries)-1]
        a.srv.LastState = &last.State
    }

    log.Println("Сервер запущен...")
    a.srv.Listen(":1312")  // блокирующий вызов
}


//
// ===== Watchdog =====
//

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
    if a.lastState == 2 {
        return
    }

    a.notifyState(2)

    now := time.Now().UTC().Add(time.Duration(config.GMT) * time.Hour)
    a.bot.UpdatePanel(fmt.Sprintf(
        "[%s] ⚠️ Потеря сигнала",
        now.Format("02.01 15:04:05"),
    ))
}
