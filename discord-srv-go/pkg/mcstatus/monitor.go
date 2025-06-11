package mcstatus

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/taiki2523/discord-srv-go/pkg/message"
	"github.com/taiki2523/discord-srv-go/pkg/notifier"
)

func getContainerHealth(containerName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "--format={{.State.Health.Status}}", containerName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	status := strings.TrimSpace(out.String())
	return status, nil
}

func StartStatusMonitor(notifier notifier.Notifier, containerName string, interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wasHealthy bool = false

	for {
		select {
		case <-ticker.C:
			status, err := getContainerHealth(containerName)
			if err != nil {
				log.Error().Err(err).Str("container", containerName).Msg("Docker healthcheck 取得失敗")
				continue
			}

			isHealthy := (status == "healthy")

			// 🔍 常時デバッグ出力
			log.Debug().
				Str("container", containerName).
				Str("status", status).
				Bool("wasHealthy", wasHealthy).
				Bool("isHealthy", isHealthy).
				Msg("Minecraft ヘルスチェック")

			// 起動検知
			if isHealthy && !wasHealthy {
				msg := message.FormatServerEvent("start")
				log.Info().Msg(msg)
				_ = notifier.Send(msg)
			}

			// 停止検知
			if !isHealthy && wasHealthy {
				msg := message.FormatServerEvent("stop")
				log.Warn().Msg(msg)
				_ = notifier.Send(msg)
			}

			wasHealthy = isHealthy

		case <-stopCh:
			return
		}
	}
}
