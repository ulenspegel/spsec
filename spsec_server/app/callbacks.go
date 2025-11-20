package app

import (
	"fmt"
	"log"
	"time"

	"spsec/config"
	"spsec/mode"
)

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
	a.srv.OnNewState = func(state int, ts int64) {
		a.mu.Lock()
		a.lastMsgTime = time.Unix(ts, 0)
		a.mu.Unlock()

		// только передаём состояние дальше в mode,
		// а сообщения отправляет notifyState!
		a.mode.OnState(state)
	}
}
