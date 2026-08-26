package live

import (
	"testing"
	"time"
)

func TestParseScheduleTime(t *testing.T) {
	rfc3339Value := "2026-08-26T10:30:00+08:00"
	rfc3339Time, err := parseScheduleTime(rfc3339Value)
	if err != nil {
		t.Fatalf("parseScheduleTime(%q) returned error: %v", rfc3339Value, err)
	}
	if got := rfc3339Time.Format(time.RFC3339); got != rfc3339Value {
		t.Fatalf("parsed RFC3339 time = %q, want %q", got, rfc3339Value)
	}

	localValue := "2026-08-26 10:30:00"
	localTime, err := parseScheduleTime(localValue)
	if err != nil {
		t.Fatalf("parseScheduleTime(%q) returned error: %v", localValue, err)
	}
	if got := localTime.Format("2006-01-02 15:04:05"); got != localValue {
		t.Fatalf("parsed local time = %q, want %q", got, localValue)
	}

	if _, err := parseScheduleTime("not-a-time"); err == nil {
		t.Fatal("parseScheduleTime accepted an invalid value")
	}
}

func TestIsValidScheduleStatus(t *testing.T) {
	for _, status := range []string{"pending", "canceled", "live"} {
		if !isValidScheduleStatus(status) {
			t.Fatalf("isValidScheduleStatus(%q) = false, want true", status)
		}
	}
	for _, status := range []string{"", "ended", "unknown"} {
		if isValidScheduleStatus(status) {
			t.Fatalf("isValidScheduleStatus(%q) = true, want false", status)
		}
	}
}
