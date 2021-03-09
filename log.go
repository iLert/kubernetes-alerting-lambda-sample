package shared

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogInit initialize log format and level
func LogInit() {
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    true,
		PartsOrder: []string{"time", "message", "caller", "level"},
		FormatLevel: func(i interface{}) string {
			return strings.ToLower(fmt.Sprintf("level=%s", i))
		},
	})
	logLevel := GetEnv("LOG_LEVEL", "info")
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Debug().Msg("Logger initialized")
}
