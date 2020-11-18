package tg

import (
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (s *Service) CallbackQueryStat(query *tgbotapi.CallbackQuery, parts []string) {
	var (
		err error
	)
	if len(parts) == 1 {
		return
	}
	args := strings.Split(parts[1], "#")
	chunk := 0
	if len(args) == 1 && args[0] == "refresh" {
		chunk = 1
	} else if len(args) == 2 && args[0] == "get" {
		chunk, err = strconv.Atoi(args[1])
		if err != nil {
			log.Printf("Error parse callback stat data '%s': %s\n", parts[1], err)
		}
	}
	if chunk != 0 {
		s.UpdateStatMessage(chunk)
	}
}
