package main

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/taiki2523/discord-srv-go/pkg/cron"
	"github.com/taiki2523/discord-srv-go/pkg/mcstatus"
	"github.com/taiki2523/discord-srv-go/pkg/notifier"
	"github.com/taiki2523/discord-srv-go/pkg/util"
	"github.com/taiki2523/discord-srv-go/pkg/watcher"
)

func main() {
	util.InitLogger()

	logPath := strings.Trim(os.Getenv("LOG_FILE"), `"`)
	webhook := strings.Trim(os.Getenv("DISCORD_WEBHOOK_URL"), `"`)
	playerCheckIntervalStr := strings.Trim(os.Getenv("PLAYER_CHECK_INTERVAL"), `"`)
	if playerCheckIntervalStr == "" {
		playerCheckIntervalStr = "5m"
	}
	playerCheckInterval, err := time.ParseDuration(playerCheckIntervalStr)
	if err != nil {
		log.Fatal().Err(err).Msg("PLAYER_CHECK_INTERVAL の形式が不正です")
	}

	util.LogStartupEnv()

	if logPath == "" || webhook == "" {
		log.Fatal().Msg("環境変数 LOG_FILE と DISCORD_WEBHOOK_URL を指定してください")
	}

	n := &notifier.DiscordNotifier{WebhookURL: webhook}
	stopCh := make(chan struct{})

	mcContainer := os.Getenv("MC_CONTAINER_NAME")
	if mcContainer == "" {
		log.Fatal().Msg("環境変数 MC_CONTAINER_NAME が設定されていません")
	}

	intervalStr := os.Getenv("MC_MONITOR_INTERVAL")
	if intervalStr == "" {
		intervalStr = "30s"
	}
	mcInterval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatal().Err(err).Msg("MC_MONITOR_INTERVAL の形式が不正です")
	}

	go mcstatus.StartStatusMonitor(n, mcContainer, mcInterval, stopCh)

	go cron.StartPlayerCheckAnnounce(n, playerCheckInterval, stopCh, logPath)

	if err := watcher.WatchFileLoop(logPath, n, stopCh); err != nil {
		log.Fatal().Err(err).Msg("アプリケーションエラー")
	}
}
