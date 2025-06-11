package cron

import (
	"time"

	"github.com/taiki2523/discord-srv-go/pkg/message"
	"github.com/taiki2523/discord-srv-go/pkg/notifier"
	"github.com/taiki2523/discord-srv-go/pkg/player"
)

func StartPlayerCheckAnnounce(n notifier.Notifier, interval time.Duration, stopCh <-chan struct{}, logPath string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			players := player.GetPlayerList(logPath)
			if len(players) == 0 {
				continue // 誰もいなければスキップ
			}

			body := message.FormatPlayerListStatus(players)
			if body == "" {
				continue
			}
			_ = n.Send(body)

		case <-stopCh:
			return
		}
	}
}
