package app

import (
    "log"
    "sync"
    "fmt"
    "time"

    "spsec/bot"
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
    // загружаем записи из диска сразу
    if err := log.LoadIntoBuffer(); err != nil {
        fmt.Println("Ошибка загрузки логов:", err)
    }

    a := &App{
        log:               log,
        srv:               srv,
        bot:               bot,
        mode:              initial,
        currentModeID:     0,
        scheduleStart:     9,
        scheduleEnd:       22,
        lastMsgTime:       time.Now(),
        timeoutSeconds:    10,
        lastState:         0,
        lastHeartbeatState: -1, // чтобы отслеживать первое сообщение
    }

    a.initCallbacks()
    return a
}


func (a *App) Run() {
    go a.bot.Start()
    go a.watchdog()

    a.srv.OnHeartbeat = a.handleHeartbeat

    // Восстановление последнего состояния из диска
    if entries, err := a.log.LoadFromDisk(); err == nil && len(entries) > 0 {
        last := entries[len(entries)-1]
        a.srv.LastState = &last.State

        // добавляем все записи в память Logger
        for _, e := range entries {
            a.log.AddEntry(e)
        }
    }

    log.Println("Сервер запущен...")
    a.srv.Listen(":1312")  // блокирующий вызов
}
