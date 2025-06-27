package utils

import (
	"log"
	"strings"
)

type LogLevel byte

const (
	Off   LogLevel = 0
	Debug LogLevel = 1 << iota
	Trace
	Info
	Error
	Critical
	All LogLevel = Debug | Trace | Info | Error | Critical
)

func (l LogLevel) String() string {

	if l == Off {
		return "OFF"
	}

	if l == All {
		return "ALL"
	}

	var strParts []string

	if l&Debug != 0 {
		strParts = append(strParts, "DEBUG")
	}

	if l&Trace != 0 {
		strParts = append(strParts, "TRACE")
	}

	if l&Info != 0 {
		strParts = append(strParts, "INFO")
	}

	if l&Error != 0 {
		strParts = append(strParts, "ERROR")
	}

	if l&Critical != 0 {
		strParts = append(strParts, "CRITICAL")
	}

	if len(strParts) == 0 {
		return "UNKNOWN"
	}

	return strings.Join(strParts, " | ")
}

type Logger struct {
	level LogLevel
}

func NewLogger() *Logger {
	return &Logger{
		level: Info,
	}
}

func (l *Logger) SetLevel(level LogLevel) *Logger {
	l.level = level
	return l
}

func (l Logger) GetLevel() LogLevel {
	return l.level
}

func (l Logger) Log(level LogLevel, message string, args ...any) {
	if (l.GetLevel() != Off && l.GetLevel() <= level) || (l.GetLevel() == All) {
		log_new_message := "[" + level.String() + "]: " + message
		log.Printf(log_new_message, args...)
	}
}

func (l Logger) Debug(message string, args ...any) {
	l.Log(Debug, message, args...)
}

func (l Logger) Info(message string, args ...any) {
	l.Log(Info, message, args...)
}

func (l Logger) Trace(message string, args ...any) {
	l.Log(Trace, message, args...)
}

func (l Logger) Error(message string, args ...any) {
	l.Log(Error, message, args...)
}

func (l Logger) Critical(message string, args ...any) {
	l.Log(Critical, message, args...)
}
