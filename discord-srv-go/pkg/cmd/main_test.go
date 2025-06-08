package main

import (
	"os"
	"testing"
)

func TestNotifyJoinLeave(t *testing.T) {
	url := os.Getenv("DISCORD_WEBHOOK_URL")
	if url == "" {
		t.Skip("DISCORD_WEBHOOK_URL が未設定のためスキップします")
	}

	notifier := &DiscordNotifier{WebhookURL: url}

	tests := []struct {
		name       string
		logLine    string
		expectSend bool
		expectMsg  string
	}{
		{
			name:       "Join event",
			logLine:    "[15:57:19] [Server thread/INFO]: marcia2525dayo joined the game",
			expectSend: true,
			expectMsg:  "🟢 marcia2525dayo がサーバに参加しました",
		},
		{
			name:       "Leave event",
			logLine:    "[15:57:22] [Server thread/INFO]: marcia2525dayo left the game",
			expectSend: true,
			expectMsg:  "🔴 marcia2525dayo がサーバから退出しました",
		},
		{
			name:       "Non-matching log",
			logLine:    "[15:53:02] [Server thread/INFO]: [Rcon: Automatic saving is now disabled]",
			expectSend: false,
		},
		{
			name:       "Malformed log",
			logLine:    "INVALID LOG LINE FORMAT",
			expectSend: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := extractPlayerName(tc.logLine)
			var message string

			if tc.expectSend {
				if player == "" {
					t.Errorf("プレイヤー名の抽出に失敗しました: %s", tc.logLine)
					return
				}
				if containsJoin(tc.logLine) {
					message = "🟢 " + player + " がサーバに参加しました"
				} else if containsLeave(tc.logLine) {
					message = "🔴 " + player + " がサーバから退出しました"
				}
				if message != tc.expectMsg {
					t.Errorf("生成されたメッセージが想定と異なります: got=%q want=%q", message, tc.expectMsg)
				}
				err := notifier.Send(message)
				if err != nil {
					t.Errorf("Discord通知に失敗: %v", err)
				}
			} else {
				if player != "" {
					t.Errorf("通知不要のログでプレイヤー名が抽出されました: %s", player)
				}
			}
		})
	}
}

// ヘルパー関数（main.goにある関数と一致させる）
func containsJoin(line string) bool {
	return len(line) > 0 && stringContains(line, "joined the game")
}
func containsLeave(line string) bool {
	return len(line) > 0 && stringContains(line, "left the game")
}
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s[len(s)-len(substr):] == substr || contains(s, substr))
}
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}
