package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/zekrotja/rogu"
	"github.com/zekrotja/rogu/level"
)

type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type RingLogger struct {
	mu    sync.RWMutex
	logs  []LogEntry
	Limit int
}

func NewRingLogger(limit int) *RingLogger {
	return &RingLogger{Limit: limit, logs: make([]LogEntry, 0, limit)}
}

func (rl *RingLogger) Write(lvl level.Level, fields []*rogu.Field, tag string, err error, errFormat string, callerFile string, callerLine int, msg string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry := LogEntry{
		Level:   fmt.Sprint(lvl),
		Message: msg,
		Time:    time.Now(),
	}

	for _, f := range fields {
		entry.Message += fmt.Sprintf(" | %s: %v", f.Key, f.Val)
	}

	if err != nil {
		entry.Message += " | error: " + err.Error()
	}

	if len(rl.logs) >= rl.Limit {
		rl.logs = rl.logs[1:]
	}
	rl.logs = append(rl.logs, entry)
	return nil
}

func (rl *RingLogger) GetLogs() []LogEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	res := make([]LogEntry, len(rl.logs))
	copy(res, rl.logs)
	return res
}

var GlobalRingLogger = NewRingLogger(1000)
