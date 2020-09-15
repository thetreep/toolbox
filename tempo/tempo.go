package tempo

import (
	"errors"
	"regexp"
	"time"

	"github.com/thetreep/toolbox/convert"
)

var (
	errISO8601DurationFormat = errors.New("bad iso8601 duration format")
	durationRgx              = regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
)

//ParseISO8601Duration converts and ISO8601 duration string to time.Duration
func ParseISO8601Duration(val string) (time.Duration, error) {
	matches := durationRgx.FindStringSubmatch(val)
	if len(matches) != 7 {
		return 0, errISO8601DurationFormat
	}

	i64 := convert.ToInt64
	years := i64(matches[1])
	months := i64(matches[2])
	days := i64(matches[3])
	hours := i64(matches[4])
	minutes := i64(matches[5])
	seconds := i64(matches[6])

	duration := time.Duration(years*24*365) * time.Hour
	duration += time.Duration(months*30*24) * time.Hour
	duration += time.Duration(days*24) * time.Hour
	duration += time.Duration(hours) * time.Hour
	duration += time.Duration(minutes) * time.Minute
	duration += time.Duration(seconds) * time.Second
	return duration, nil
}
