package message

import (
	"fmt"
	"strings"

	"github.com/taiki2523/discord-srv-go/pkg/util"
)

const delimiter = "===================="

// サーバ起動・停止イベント
func FormatServerEvent(eventType string) string {
	timestamp := util.GetNow().Format("2006-01-02 15:04:05")
	var body string

	switch eventType {
	case "start":
		body = "🟢🖥️ Minecraftサーバが起動しました"
	case "stop":
		body = "🔴🖥️ Minecraftサーバが停止しました"
	default:
		return ""
	}

	return fmt.Sprintf("%s\n%s\n\n📅 発生時刻: %s\n%s", delimiter, body, timestamp, delimiter)
}

// プレイヤー参加・退出イベント
func FormatPlayerEvent(eventType, playerName string) string {
	timestamp := util.GetNow().Format("2006-01-02 15:04:05")
	boldName := fmt.Sprintf("**%s**", playerName)

	var body string
	switch eventType {
	case "join":
		body = fmt.Sprintf("🟢👤 %s がサーバに参加しました", boldName)
	case "leave":
		body = fmt.Sprintf("🔴👤 %s がサーバから退出しました", boldName)
	default:
		return ""
	}

	return fmt.Sprintf("%s\n%s\n\n📅 発生時刻: %s\n%s", delimiter, body, timestamp, delimiter)
}

// 現在の参加者ステータス
func FormatPlayerListStatus(players []string) string {
	timestamp := util.GetNow().Format("2006-01-02 15:04:05")
	if len(players) == 0 {
		return ""
	}

	boldPlayers := make([]string, len(players))
	for i, name := range players {
		boldPlayers[i] = fmt.Sprintf("**%s**", name)
	}

	body := fmt.Sprintf("👥 現在の参加者: %s", strings.Join(boldPlayers, ", "))

	return fmt.Sprintf("%s\n%s\n\n📅 通知時刻: %s\n%s", delimiter, body, timestamp, delimiter)
}
