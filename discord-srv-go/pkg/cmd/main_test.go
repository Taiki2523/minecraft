// file: pkg/cmd/main_test.go
package main

import (
	"os"
	"strings"
	"testing"
)

// mockNotifier is used for verifying notifications in tests
type mockNotifier struct {
	Messages []string
}

func (m *mockNotifier) Send(msg string) error {
	m.Messages = append(m.Messages, msg)
	return nil
}

func TestProcessLogLine(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"[Server thread/INFO]: Steve joined the game", "🟢 Steve がサーバに参加しました"},
		{"[Server thread/INFO]: Alex left the game", "🔴 Alex がサーバから退出しました"},
		{"[Server thread/INFO]: unrelated message", ""},
	}

	for _, tt := range tests {
		mock := &mockNotifier{}
		processLogLine(tt.line, mock)
		if tt.expected == "" && len(mock.Messages) > 0 {
			t.Errorf("unexpected message: %v", mock.Messages)
		}
		if tt.expected != "" && (len(mock.Messages) != 1 || mock.Messages[0] != tt.expected) {
			t.Errorf("got %v, want %v", mock.Messages, tt.expected)
		}
	}
}

func TestRunWithNotifier_FileNotFound(t *testing.T) {
	mock := &mockNotifier{}
	err := RunWithNotifier("nonexistent.log", mock, 2)
	if err == nil || !strings.Contains(err.Error(), "failed to open") {
		t.Errorf("expected file open failure, got: %v", err)
	}
}

// Integration test for actual Discord Webhook
func TestDiscordNotification_Integration(t *testing.T) {
	webhook := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhook == "" {
		t.Skip("DISCORD_WEBHOOK_URL is not set")
	}

	notifier := &DiscordNotifier{WebhookURL: webhook}
	err := notifier.Send("✅ Integration test from descord-srv-go")
	if err != nil {
		t.Fatalf("Failed to send Discord message: %v", err)
	}
}
