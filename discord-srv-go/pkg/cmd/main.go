// file: pkg/cmd/main.go
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Notifier defines interface for sending messages
type Notifier interface {
	Send(message string) error
}

// DiscordNotifier sends messages to Discord Webhook
type DiscordNotifier struct {
	WebhookURL string
}

func (d *DiscordNotifier) Send(message string) error {
	payload := map[string]string{"content": message}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(d.WebhookURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response from Discord: %s", resp.Status)
	}
	return nil
}

func extractPlayerName(line string) string {
	parts := strings.Split(line, "]: ")
	if len(parts) < 2 {
		return ""
	}
	fields := strings.Fields(parts[1])
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

func formatEventMessage(icon string, name string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s %s %s\n\n発生時刻: %s", icon, name, eventText(icon), timestamp)
}

func eventText(icon string) string {
	switch icon {
	case "🟢":
		return "がサーバに参加しました"
	case "🔴":
		return "がサーバから退出しました"
	default:
		return ""
	}
}

func getListFilePath(baseLogPath string) string {
	dir := "/data/.descord-srv"
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "active_players.txt")
}

func updatePlayerList(baseLogPath, name string, joined bool) {
	listFile := getListFilePath(baseLogPath)
	players := make(map[string]struct{})

	if file, err := os.Open(listFile); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			players[scanner.Text()] = struct{}{}
		}
		file.Close()
	}

	if joined {
		players[name] = struct{}{}
	} else {
		delete(players, name)
	}

	if err := os.MkdirAll(filepath.Dir(listFile), 0755); err != nil {
		log.Error().Err(err).Msg("プレイヤーリストディレクトリの作成失敗")
		return
	}

	file, err := os.Create(listFile)
	if err != nil {
		log.Error().Err(err).Msg("プレイヤーリストの更新失敗")
		return
	}
	defer file.Close()
	for p := range players {
		fmt.Fprintln(file, p)
	}
}

func getPlayerList(baseLogPath string) []string {
	listFile := getListFilePath(baseLogPath)
	players := []string{}

	if file, err := os.Open(listFile); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			players = append(players, scanner.Text())
		}
		file.Close()
	}
	return players
}

func processLogLine(line string, notifier Notifier, logPath string) {
	log.Debug().Str("line", line).Msg("Checking log line")

	if strings.Contains(line, "joined the game") {
		if name := extractPlayerName(line); name != "" {
			updatePlayerList(logPath, name, true)
			msg := formatEventMessage("🟢", name)
			err := notifier.Send(msg)
			if err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	} else if strings.Contains(line, "left the game") {
		if name := extractPlayerName(line); name != "" {
			updatePlayerList(logPath, name, false)
			msg := formatEventMessage("🔴", name)
			err := notifier.Send(msg)
			if err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	}
}

func watchFileLoop(logPath string, notifier Notifier, stopCh <-chan struct{}) error {
	var file *os.File
	var reader *bufio.Reader
	var err error
	var currentInode uint64

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dir := filepath.Dir(logPath)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	openLogFile := func() error {
		f, err := os.Open(logPath)
		if err != nil {
			return err
		}
		stat, _ := f.Stat()
		if stat == nil {
			return errors.New("cannot stat log file")
		}
		sysStat := stat.Sys().(*syscall.Stat_t)
		inode := sysStat.Ino
		if file != nil && inode == currentInode {
			f.Close()
			return nil
		}
		if file != nil {
			file.Close()
		}
		file = f
		reader = bufio.NewReader(file)
		file.Seek(0, io.SeekEnd)
		currentInode = inode
		log.Info().Str("file", logPath).Msg("ログファイルをオープンしました")
		return nil
	}

	_ = openLogFile()

	for {
		select {
		case <-stopCh:
			if file != nil {
				file.Close()
			}
			return nil
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) != 0 && filepath.Base(event.Name) == filepath.Base(logPath) {
				log.Debug().Str("event", event.String()).Msg("fsnotifyイベント検出")
				_ = openLogFile()
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
							time.Sleep(100 * time.Millisecond)
							break
						}
						return err
					}
					processLogLine(line, notifier, logPath)
				}
			}
		case err := <-watcher.Errors:
			log.Error().Err(err).Msg("watcher error")
		}
	}
}

func startHealthCheck(notifier Notifier, interval time.Duration, stopCh <-chan struct{}, logPath string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			timeStr := time.Now().Format("2006-01-02 15:04:05")
			players := getPlayerList(logPath)
			body := "✅ サーバは稼働中です\n\nチェック時刻: " + timeStr
			if len(players) > 0 {
				body += "\n\n現在の参加者: " + strings.Join(players, ", ")
			} else {
				body += "\n\n現在サーバには誰もいません"
			}
			err := notifier.Send(body)
			if err != nil {
				log.Error().Err(err).Msg("ヘルスチェック送信失敗")
			}
		case <-stopCh:
			return
		}
	}
}

func main() {
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

	logPath := strings.Trim(os.Getenv("LOG_FILE"), `"`)
	webhook := strings.Trim(os.Getenv("DISCORD_WEBHOOK_URL"), `"`)
	healthIntervalStr := strings.Trim(os.Getenv("HEALTH_INTERVAL"), `"`)
	if healthIntervalStr == "" {
		healthIntervalStr = "5m"
	}
	healthInterval, err := time.ParseDuration(healthIntervalStr)
	if err != nil {
		log.Fatal().Err(err).Msg("HEALTH_INTERVAL の形式が不正です")
	}

	log.Info().Msg("=== 起動時環境変数 ===")
	log.Info().Str("LOG_FILE", logPath).Send()
	log.Info().Str("DISCORD_WEBHOOK_URL", webhook).Send()
	log.Info().Str("LOG_LEVEL", logLevel).Send()
	log.Info().Str("HEALTH_INTERVAL", healthInterval.String()).Send()

	if logPath == "" || webhook == "" {
		log.Fatal().Msg("環境変数 LOG_FILE と DISCORD_WEBHOOK_URL を指定してください")
	}

	n := &DiscordNotifier{WebhookURL: webhook}
	stopCh := make(chan struct{})

	go startHealthCheck(n, healthInterval, stopCh, logPath)

	if err := watchFileLoop(logPath, n, stopCh); err != nil {
		log.Fatal().Err(err).Msg("アプリケーションエラー")
	}
}
