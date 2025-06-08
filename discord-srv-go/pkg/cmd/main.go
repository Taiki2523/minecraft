// file: pkg/cmd/main.go
package main

import (
	"bufio"
	"errors"
	"flag"
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
	logDebug("Sending message to Discord: %s", message)
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

var logLevel = "info"

func logDebug(format string, v ...interface{}) {
	if logLevel == "debug" {
		log.Printf("🐛 "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	log.Printf("ℹ️ "+format, v...)
}

func logError(format string, v ...interface{}) {
	log.Printf("❌ "+format, v...)
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

func processLogLine(line string, notifier Notifier) {
	logDebug("Checking line: %s", line)

	if strings.Contains(line, "joined the game") {
		if name := extractPlayerName(line); name != "" {
			err := notifier.Send(fmt.Sprintf("🟢 %s がサーバに参加しました", name))
			if err != nil {
				logError("通知失敗: %v", err)
			}
		}
	} else if strings.Contains(line, "left the game") {
		if name := extractPlayerName(line); name != "" {
			err := notifier.Send(fmt.Sprintf("🔴 %s がサーバから退出しました", name))
			if err != nil {
				logError("通知失敗: %v", err)
			}
		}
	}
}

func watchFile(logPath string, notifier Notifier) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dir := filepath.Dir(logPath)
	logInfo("Watching directory: %s", dir)

	err = watcher.Add(logPath)
	if err != nil {
		return err
	}

	file, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(file)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("fsnotify event channel closed")
			}
			if event.Op&fsnotify.Write == fsnotify.Write && event.Name == logPath {
				logInfo("🔁 Log file modified: %s", event.Name)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return err
					}
					processLogLine(line, notifier)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("fsnotify error channel closed")
			}
			logError("Watcher error: %v", err)
		}
	}
}

func RunWithNotifier(logPath string, notifier Notifier, maxAttempts int) error {
	logInfo("RunWithNotifier started. Monitoring: %s", logPath)

	for i := 1; i <= maxAttempts; i++ {
		if _, err := os.Stat(logPath); err == nil {
			return watchFile(logPath, notifier)
		}
		logInfo("Log file not found. Retrying in 5 seconds... (%d/%d)", i, maxAttempts)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("failed to find log file after %d retries", maxAttempts)
}

func main() {
	logPath := flag.String("log", "/data/logs/latest.log", "Path to the Minecraft log file")
	webhook := flag.String("webhook", os.Getenv("DISCORD_WEBHOOK_URL"), "Discord Webhook URL")
	retries := flag.Int("retries", 12, "Max retries if log file not found")
	level := flag.String("loglevel", "info", "Log level: debug, info")
	flag.Parse()

	if *webhook == "" {
		log.Fatal("DISCORD_WEBHOOK_URL is not set or --webhook flag not provided")
	}
	logLevel = *level

	n := &DiscordNotifier{WebhookURL: *webhook}
	if err := RunWithNotifier(*logPath, n, *retries); err != nil {
		log.Fatal(err)
	}
}
