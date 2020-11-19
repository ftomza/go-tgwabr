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

func (s *Service) CallbackQuerySomethingElse(query *tgbotapi.CallbackQuery, parts []string) {
	if len(parts) == 1 {
		return
	}
	args := strings.Split(parts[1], "#")
	if args[0] != "get" {
		return
	}
	msg := query.Message
	msg.From = query.From
	s.CommandSomethingElse(tgbotapi.Update{Message: msg}, args[1], args[2])
}

func (s *Service) CallbackQueryChat(update tgbotapi.Update, parts []string) {
	if len(parts) == 1 {
		return
	}
	args := strings.Split(parts[1], "#")
	if args[0] != "join" {
		return
	}
	msg := update.CallbackQuery.Message
	msg.From = update.CallbackQuery.From
	_, _ = s.bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Join "+args[1]))
	s.CommandJoin(tgbotapi.Update{Message: msg}, args[1], args[2])
}
