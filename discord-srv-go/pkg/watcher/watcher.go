package watcher

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/taiki2523/discord-srv-go/pkg/message"
	"github.com/taiki2523/discord-srv-go/pkg/notifier"
	"github.com/taiki2523/discord-srv-go/pkg/player"
)

func WatchFileLoop(logPath string, notifier notifier.Notifier, stopCh <-chan struct{}) error {
	var file *os.File
	var reader *bufio.Reader
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
			return errors.New("ログファイルの stat に失敗")
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

func processLogLine(line string, notifier notifier.Notifier, logPath string) {
	log.Debug().Str("line", line).Msg("Checking log line")

	if strings.Contains(line, "joined the game") {
		if name := player.ExtractPlayerName(line); name != "" {
			player.UpdatePlayerList(logPath, name, true)
			msg := message.FormatPlayerEvent("join", name)
			if err := notifier.Send(msg); err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	} else if strings.Contains(line, "left the game") {
		if name := player.ExtractPlayerName(line); name != "" {
			player.UpdatePlayerList(logPath, name, false)
			msg := message.FormatPlayerEvent("leave", name)
			if err := notifier.Send(msg); err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	}
}
