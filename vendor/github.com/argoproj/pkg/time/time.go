package time

import (
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var durationRegex = regexp.MustCompile(`^(\d+)([smhd])$`)

// ParseDuration parses a duration string and returns the time.Duration
func ParseDuration(duration string) (*time.Duration, error) {
	matches := durationRegex.FindStringSubmatch(duration)
	if len(matches) != 3 {
		return nil, errors.Errorf("Invalid since format '%s'. Expected format <duration><unit> (e.g. 3h)\n", duration)
	}
	amount, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	var unit time.Duration
	switch matches[2] {
	case "s":
		unit = time.Second
	case "m":
		unit = time.Minute
	case "h":
		unit = time.Hour
	case "d":
		unit = time.Hour * 24
	}
	dur := unit * time.Duration(amount)
	return &dur, nil
}

// ParseSince parses a duration string and returns a time.Time in history relative to current time
func ParseSince(duration string) (*time.Time, error) {
	dur, err := ParseDuration(duration)
	if err != nil {
		return nil, err
	}
	since := time.Now().UTC().Add(-*dur)
	return &since, nil
}
