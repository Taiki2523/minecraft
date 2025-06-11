package player

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func getListFilePath(basePath string) string {
	dir := "/tmp/discord-srv-go"
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "active_players.txt")
}

func UpdatePlayerList(basePath, name string, joined bool) {
	listFile := getListFilePath(basePath)
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

	file, err := os.Create(listFile)
	if err != nil {
		return
	}
	defer file.Close()
	for p := range players {
		fmt.Fprintln(file, p)
	}
}

func GetPlayerList(basePath string) []string {
	listFile := getListFilePath(basePath)
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
