package workflow

import "time"

func isValidTimeStamp(candidate string, format string) bool {
	_, err := time.Parse(format, candidate)
	if err != nil {
		return false
	}
	return true
}

// isValidDateInput tries to parse the time string with multiple formats\
// and returns true if the time string is valid
func isValidDateInput(candidate string) bool {
	timeFormats := [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.Kitchen,
		time.RFC3339,
		time.RFC3339Nano,
		time.DateTime,
		time.DateOnly,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.TimeOnly,
		"2006-01-02",                         // RFC 3339
		"2006-01-02 15:04",                   // RFC 3339 with minutes
		"2006-01-02 15:04:05",                // RFC 3339 with seconds
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds; replace Z with either + or -. use Z for UTC
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds; replace Z with either + or -. use Z for UTC
	}
	for _, format := range timeFormats {
		if isValidTimeStamp(candidate, format) {
			return true
		}
	}
	return false
}
