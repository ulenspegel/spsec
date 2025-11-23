package serv

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var digitRegex = regexp.MustCompile(`^[0-9]$`)

// Server хранит внутреннее состояние и колбэки
type Server struct {
	LastState   *int
	OnNewState  func(state int, ts int64) // вызывается только при смене состояния
	OnHeartbeat func(ts int64)            // вызывается при каждом приходе сообщения
}

// NewServer создает сервер
func NewServer() *Server {
	return &Server{}
}

// HandleDigit обрабатывает HTTP POST с одной цифрой
func (s *Server) HandleDigit(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	digit := string(body)
	if !digitRegex.MatchString(digit) {
		http.Error(w, "body must contain exactly one digit 0-9", http.StatusBadRequest)
		return
	}

	state, _ := strconv.Atoi(digit)
	ts := time.Now().Unix()

	// Heartbeat всегда
	if s.OnHeartbeat != nil {
		s.OnHeartbeat(ts)
	}

	// Проверяем, был ли сигнал ранее потерян
	wasNoSignal := s.LastState != nil && *s.LastState == 2

	// Если состояние изменилось — вызываем OnNewState
	if s.LastState == nil || *s.LastState != state {
		if s.OnNewState != nil {
			s.OnNewState(state, ts)
		}
		s.LastState = &state
		log.Printf("New state received: %d", state)
	} else if wasNoSignal && (state == 0 || state == 1) {
		// Сигнал восстановлен, даже если state не изменился
		if s.OnNewState != nil {
			s.OnNewState(state, ts)
		}
		s.LastState = &state
		log.Printf("Signal restored: %d", state)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK: %s\n", digit)
}

// Listen запускает HTTP-сервер на указанном порту
func (s *Server) Listen(port string) error {
	http.HandleFunc("/", s.HandleDigit)
	log.Printf("Listening on %s...", port)
	return http.ListenAndServe(port, nil)
}
