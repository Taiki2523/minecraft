// file: pkg/cmd/main.go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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
	log.Println("📤 Sending message to Discord:", message)
	payload := strings.NewReader(`{"content":"` + message + `"}`)
	resp, err := http.Post(d.WebhookURL, "application/json", payload)
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
	msg := parts[1]
	fields := strings.Fields(msg)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

func processLogLine(line string, notifier Notifier) {
	log.Printf("📄 解析中ログ行: %s", line)

	if strings.Contains(line, "joined the game") {
		if name := extractPlayerName(line); name != "" {
			err := notifier.Send(fmt.Sprintf("🟢 %s がサーバに参加しました", name))
			if err != nil {
				log.Printf("❌ 通知失敗: %v", err)
			}
		}
	} else if strings.Contains(line, "left the game") {
		if name := extractPlayerName(line); name != "" {
			err := notifier.Send(fmt.Sprintf("🔴 %s がサーバから退出しました", name))
			if err != nil {
				log.Printf("❌ 通知失敗: %v", err)
			}
		}
	}
}

func RunWithNotifier(logPath string, notifier Notifier, maxAttempts int) error {
	log.Println("🏁 RunWithNotifier started. Monitoring:", logPath)

	// retry open file
	var file *os.File
	var err error
	for i := 1; i <= maxAttempts; i++ {
		file, err = os.Open(logPath)
		if err == nil {
			break
		}
		log.Printf("⚠️ Log file not found. Retrying in 5 seconds... (%d/%d)", i, maxAttempts)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return errors.New("failed to open log file after max retries")
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}
	reader := bufio.NewReader(file)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dir := filepath.Dir(logPath)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write && event.Name == logPath {
				log.Println("🔁 Log file modified:", event.Name)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return fmt.Errorf("read error: %w", err)
					}
					processLogLine(line, notifier)
				}
			}
		case err := <-watcher.Errors:
			log.Println("❌ Watcher error:", err)
		}
	}
}

func main() {
	logPath := os.Getenv("MINECRAFT_LOG_PATH")
	webhook := os.Getenv("DISCORD_WEBHOOK_URL")

	if logPath == "" || webhook == "" {
		log.Fatal("❗ 環境変数が不足しています: MINECRAFT_LOG_PATH, DISCORD_WEBHOOK_URL")
	}

	notifier := &DiscordNotifier{WebhookURL: webhook}
	if err := RunWithNotifier(logPath, notifier, 10); err != nil {
		log.Fatalf("❌ Run failed: %v", err)
	}
}
