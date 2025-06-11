package util

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := strings.ToLower(strings.Trim(os.Getenv("LOG_LEVEL"), `"`))
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info", "":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		log.Fatal().Msg("LOG_LEVEL は debug または info を指定してください")
	}
}

func LogStartupEnv() {
	log.Info().Msg("=== 起動時環境変数 ===")
	log.Info().Str("LOG_FILE", strings.Trim(os.Getenv("LOG_FILE"), `"`)).Send()
	log.Info().Str("DISCORD_WEBHOOK_URL", strings.Trim(os.Getenv("DISCORD_WEBHOOK_URL"), `"`)).Send()
	log.Info().Str("LOG_LEVEL", strings.Trim(os.Getenv("LOG_LEVEL"), `"`)).Send()
	log.Info().Str("PLAYER_CHECK_INTERVAL", strings.Trim(os.Getenv("PLAYER_CHECK_INTERVAL"), `"`)).Send()
	log.Info().Str("MC_CONTAINER_NAME", strings.Trim(os.Getenv("MC_CONTAINER_NAME"), `"`)).Send()
	log.Info().Str("MC_MONITOR_INTERVAL", strings.Trim(os.Getenv("MC_MONITOR_INTERVAL"), `"`)).Send()
	log.Info().Str("TIMEZONE", strings.Trim(os.Getenv("TIMEZONE"), `"`)).Send()
}
