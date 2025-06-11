package player

import (
	"fmt"
	"strings"

	"github.com/taiki2523/discord-srv-go/pkg/util"
)

func ExtractPlayerName(line string) string {
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

func FormatEventMessage(icon, name string) string {
	timestamp := util.GetNow().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s %s %s\n\n発生時刻: %s", icon, name, EventText(icon), timestamp)
}

func EventText(icon string) string {
	switch icon {
	case "🟢":
		return "がサーバに参加しました"
	case "🔴":
		return "がサーバから退出しました"
	default:
		return ""
	}
}
