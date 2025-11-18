package mode

import (
	"time"
	// "spsec/config"
)


// Mode — интерфейс для всех режимов работы
type Mode interface {
	// OnState — вызывается при изменении состояния
	OnState(state int)

	// Name — имя режима (для логов или бота)
	Name() string
}


type SilentMode struct{}

func (SilentMode) OnState(state int) {
	// ничего не делаем
}

func (SilentMode) Name() string {
	return "silent"
}

func NewSilent() Mode {
	return SilentMode{}
}


// func stateToString(state int) string {
// 	switch state {
// 	case 0:
// 		return "Закрыто"
// 	case 1:
// 		return "Открыто"
// 	case 2:
// 		return "Не в сети"
// 	default:
// 		return "Бля"
// 	}
// }

type Broadcaster interface {
	Broadcast(msg string)
}


type AlarmMode struct {
	sendFn func(state int)
}

func NewAlarm(sendFn func(state int)) Mode {
	return &AlarmMode{sendFn: sendFn}
}

func (a *AlarmMode) OnState(state int) {
	if a.sendFn != nil {
		a.sendFn(state)
	}
}

func (a *AlarmMode) Name() string {
	return "alarm"
}











type ScheduleMode struct {
	startHour int
	endHour   int
	alarm     Mode
	silent    Mode
}

func NewSchedule(start, end int, alarmFn func(int)) Mode {
	return &ScheduleMode{
		startHour: start,
		endHour:   end,
		alarm:     NewAlarm(alarmFn),
		silent:    NewSilent(),
	}
}

func (m *ScheduleMode) OnState(state int) {
	now := time.Now().Hour()

	active := now >= m.startHour && now < m.endHour

	if active {
		m.alarm.OnState(state)
	} else {
		m.silent.OnState(state)
	}
}

func (m *ScheduleMode) Name() string {
	return "schedule"
}