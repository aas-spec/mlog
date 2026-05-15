package mlog

import (
	"regexp"
	"testing"
)

func TestGetTimeStamp(t *testing.T) {
	ts := GetTimeStamp()
	matched, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, ts)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("unexpected timestamp format: %s", ts)
	}
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel("test-level", 3)
	logger := getLogger("test-level")
	if logger.Level != 3 {
		t.Fatalf("expected Level=3, got %d", logger.Level)
	}
}

func TestSetStoreDays(t *testing.T) {
	SetStoreDays("test-store", 14)
	logger := getLogger("test-store")
	if logger.StoreDays != 14 {
		t.Fatalf("expected StoreDays=14, got %d", logger.StoreDays)
	}
}

func TestSetLogUseOwnDir(t *testing.T) {
	SetLogUseOwnDir("test-dir", true)
	logger := getLogger("test-dir")
	if !logger.UseOwnDir {
		t.Fatal("expected UseOwnDir=true")
	}
}

func TestPanicArgs(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		s, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic string, got %T: %v", r, r)
		}
		if s == "" {
			t.Fatal("panic message is empty")
		}
	}()
	defLogger.Panic("something went wrong")
}
