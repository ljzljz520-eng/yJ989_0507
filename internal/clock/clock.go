package clock

import "time"

type Clock interface {
	Now() time.Time
}

type FixedClock struct {
	current time.Time
}

func NewFixed(start time.Time) *FixedClock {
	return &FixedClock{current: start.UTC()}
}

func (clock *FixedClock) Now() time.Time {
	return clock.current
}

func (clock *FixedClock) Advance(duration time.Duration) {
	clock.current = clock.current.Add(duration)
}

func (clock *FixedClock) Set(value time.Time) {
	clock.current = value.UTC()
}

func At(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}
