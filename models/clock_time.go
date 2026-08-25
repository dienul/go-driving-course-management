package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ClockTime maps PostgreSQL TIME columns without attaching a calendar date.
type ClockTime struct{ time.Time }

func NewClockTime(value time.Time) ClockTime { return ClockTime{Time: value} }

func (c *ClockTime) Scan(value any) error {
	switch input := value.(type) {
	case time.Time:
		c.Time = input
		return nil
	case string:
		return c.parse(input)
	case []byte:
		return c.parse(string(input))
	default:
		return fmt.Errorf("cannot scan %T into ClockTime", value)
	}
}

func (c *ClockTime) parse(input string) error {
	for _, layout := range []string{"15:04:05.999999999", "15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, input)
		if err == nil {
			c.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("invalid PostgreSQL TIME value %q", input)
}

func (c ClockTime) Value() (driver.Value, error) {
	return c.Time.Format("15:04:05"), nil
}

func (c ClockTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Time.Format("15:04:05"))
}

func (c *ClockTime) UnmarshalJSON(input []byte) error {
	var value string
	if err := json.Unmarshal(input, &value); err != nil {
		return err
	}
	return c.parse(value)
}
