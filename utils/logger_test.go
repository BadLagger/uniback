package utils

import (
	"testing"
)

func TestLoggerLevelString(t *testing.T) {
	if Off.String() != "OFF" {
		t.Errorf("Expected Off == String(OFF), but %s", Off.String())
	}

	if Debug.String() != "DEBUG" {
		t.Errorf("Expected DEBUG == String(DEBUG), but %s", Debug.String())
	}

	if Trace.String() != "TRACE" {
		t.Errorf("Expected TRACE == String(TRACE), but %s", Trace.String())
	}

	if Info.String() != "INFO" {
		t.Errorf("Expected Info == String(INFO), but %s", Info.String())
	}

	if Error.String() != "ERROR" {
		t.Errorf("Expected Error == String(ERROR), but %s", Error.String())
	}

	if Critical.String() != "CRITICAL" {
		t.Errorf("Expected Critical == String(CRITICAL), but %s", Critical.String())
	}

	if All.String() != "ALL" {
		t.Errorf("Expected All == String(ALL), but %s", All.String())
	}

	test := Debug | Critical

	if test.String() != "DEBUG | CRITICAL" {
		t.Errorf("Expected DEBUG | CRITICAL, but %s", test.String())
	}

	test = (1 << 6)

	if test.String() != "UNKNOWN" {
		t.Errorf("Expected (1 << 6) = UNKNOWN, but %s", test.String())
	}

	test = (1 << 7)

	if test.String() != "UNKNOWN" {
		t.Errorf("Expected (1 << 7) = UNKNOWN, but %s", test.String())
	}
}

func TestCreateLogger(t *testing.T) {
	logger := NewLogger()

	if logger.GetLevel() != Info {
		t.Errorf("Default Logger should be with Info level but it is: %s", logger.GetLevel().String())
	}

	logger = NewLogger().SetLevel(Debug | Error)

	if logger.GetLevel() != (Debug | Error) {
		t.Errorf("Create Logger with Debug and Errors level but it is: %s", logger.GetLevel().String())
	}
}
