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

	// ★ фикс: timestamp реального прихода
	ts := time.Now().Unix()

	// ★ фикс: присылаем heartbeat при каждом реальном сообщении
	if s.OnHeartbeat != nil {
		s.OnHeartbeat(ts)
	}

	// Если состояние изменилось — отправляем OnNewState
	if s.LastState == nil || *s.LastState != state {
		if s.OnNewState != nil {
			s.OnNewState(state, ts)
		}

		// обновляем last state
		s.LastState = &state
		log.Printf("New state received: %d", state)
	}

	// ответ клиенту
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK: %s\n", digit)
}

// Listen запускает HTTP-сервер на указанном порту
func (s *Server) Listen(port string) error {
	http.HandleFunc("/", s.HandleDigit)
	log.Printf("Listening on %s...", port)
	return http.ListenAndServe(port, nil)
}
