package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestClockTimePostgreSQLScanAndValue(t *testing.T) {
	cases := []any{"08:00:00", []byte("09:00:00.000000"), time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)}
	for _, item := range cases {
		var clock ClockTime
		if err := clock.Scan(item); err != nil {
			t.Fatalf("scan %v: %v", item, err)
		}
		value, err := clock.Value()
		if err != nil || len(value.(string)) != 8 {
			t.Fatalf("invalid SQL value %v: %v", value, err)
		}
	}
	var invalid ClockTime
	if err := invalid.Scan(123); err == nil {
		t.Fatal("unsupported SQL value type was accepted")
	}
	if err := invalid.Scan("bad-time"); err == nil {
		t.Fatal("malformed PostgreSQL TIME value was accepted")
	}
}

func TestClockTimeJSONRoundTrip(t *testing.T) {
	original := NewClockTime(time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC))
	encoded, err := json.Marshal(original)
	if err != nil || string(encoded) != `"14:00:00"` {
		t.Fatalf("marshal clock: %s, %v", encoded, err)
	}
	var decoded ClockTime
	if err := json.Unmarshal(encoded, &decoded); err != nil || decoded.Hour() != 14 {
		t.Fatalf("unmarshal clock: %+v, %v", decoded, err)
	}
}
