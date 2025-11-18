package logger

import (
	"encoding/gob"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Time  time.Time
	State int
}

// Logger — кольцевой лог с дозаписью на диск
type Logger struct {
	mu     sync.Mutex
	buffer []LogEntry
	head   int
	full   bool
	logFile string
	maxSize int64
}

// NewLogger создает новый лог
// size — размер кольцевого буфера в памяти
// logFile — путь к файлу
// maxSize — максимальный размер файла на диске
func NewLogger(size int, logFile string, maxSize int64) *Logger {
	return &Logger{
		buffer:  make([]LogEntry, size),
		logFile: logFile,
		maxSize: maxSize,
	}
}

// Add добавляет запись в лог
func (l *Logger) Add(state int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{Time: time.Now(), State: state}
	l.buffer[l.head] = entry
	l.head = (l.head + 1) % len(l.buffer)
	if l.head == 0 {
		l.full = true
	}

	return l.appendToDisk(entry)
}

// appendToDisk дозаписывает запись на диск (если файл < maxSize)
func (l *Logger) appendToDisk(entry LogEntry) error {
	info, err := os.Stat(l.logFile)
	if err == nil && info.Size() >= l.maxSize {
		return nil
	}

	f, err := os.OpenFile(l.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(entry)
}

// Last возвращает последние n записей из кольцевого буфера
func (l *Logger) Last(n int) []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	size := len(l.buffer)
	var count int
	if l.full {
		count = size
	} else {
		count = l.head
	}

	if n > count {
		n = count
	}

	res := make([]LogEntry, n)
	for i := 0; i < n; i++ {
		idx := (l.head - n + i + size) % size
		res[i] = l.buffer[idx]
	}
	return res
}

// LoadFromDisk загружает все записи из файла
func (l *Logger) LoadFromDisk() ([]LogEntry, error) {
	f, err := os.Open(l.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []LogEntry
	dec := gob.NewDecoder(f)
	for {
		var e LogEntry
		if err := dec.Decode(&e); err != nil {
			break
		}
		entries = append(entries, e)
	}
	return entries, nil
}
