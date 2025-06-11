package util

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

var location *time.Location

func init() {
	tz := os.Getenv("TIMEZONE")
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Warn().Str("TIMEZONE", tz).Msg("指定されたタイムゾーンが無効です。UTCにフォールバックします")
		loc = time.UTC
	}
	location = loc
}

func GetNow() time.Time {
	return time.Now().In(location)
}
