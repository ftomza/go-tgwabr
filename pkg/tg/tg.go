package tg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"tgwabr/api"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Service struct {
	ctx        context.Context
	bot        *tgbotapi.BotAPI
	mainGroups []int64
	api.TG
}

func New(ctx context.Context) (service *Service, err error) {

	service = &Service{ctx: ctx}

	// return nil, nil

	mainGroupsStr := os.Getenv("TG_MAIN_GROUPS")
	for _, v := range strings.Split(mainGroupsStr, ",") {
		g, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return service, fmt.Errorf("error parse ID: %w", err)
		}
		service.mainGroups = append(service.mainGroups, g)
	}

	service.bot, err = tgbotapi.NewBotAPI(os.Getenv("TG_API_TOKEN"))
	if err != nil {
		return
	}

	if os.Getenv("TG_DEBUG") != "" {
		service.bot.Debug = true
	}

	log.Printf("Authorized on account %s", service.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	err = service.bot.SetMyCommands([]tgbotapi.BotCommand{
		{"check_client", "Check possibility to join WhatsApp client, e.g. /check_client +971 55 995 02 03"},
		{"alias", "Set alias to WhatsApp client, e.g. /alias +971 55 995 02 03 Maxim"},
		{"contact", "Add contact to bot, e.g. /contact +971 55 995 02 03 Maxim"},
		{"join", "Join chat with WhatsApp client, e.g. /join +7(911) 113-59-00 minsk or /join Maxim dubai"},
		{"history", "Show recent messages (by default 10 ones) from chat with WhatsApp client, e.g. /history or /history 20"},
		{"leave", "Leave chat"},
		{"status", "Show connection status of Telegram main group to WhatsApp account"},
		{"login", "Login to definite WhatsApp account"},
		{"set", "Set Telegram main group name, e.g. /set dubai"},
		{"logout", "Logout WhatsApp account"},
		{"sync", "Try sync address book and chats(only stat)"},
		{"restart", "Restart bot"},
		{"repined", "Restore statistics in a pin"},
		{"somethingelse", "Add keyboard for fast call join chat"},
		{"autoreply", "Set auto reply to incoming messages from not joined WhatsApp client, e.g. /autoreply all \"Autoreply text here\" or /autoreply +971 55 995 02 03 \"Autoreply text here\""},
	})
	if err != nil {
		return
	}

	updates, err := service.bot.GetUpdatesChan(u)
	if err != nil {
		return
	}

	go service.mainLoop(updates)

	return
}

func (s *Service) UpdateCTX(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) ShutDown() error {
	s.bot.StopReceivingUpdates()
	return nil
}

type Message tgbotapi.Message

type APITGMessage api.TGMessage

func (a Message) ToAPIMessage() *api.TGMessage {
	item := &api.TGMessage{
		ChatID:       a.Chat.ID,
		UserName:     a.From.UserName,
		MessageID:    a.MessageID,
		Timestamp:    a.Date,
		FwdMessageID: a.ForwardFromMessageID,
	}
	return item
}

func (s *Service) BotSend(msg tgbotapi.Chattable) (response tgbotapi.Message, err error) {

	tooMany := "Too Many Requests: retry after "

	for {
		response, err = s.bot.Send(msg)
		if err == nil {
			break
		}
		var tgErr *tgbotapi.Error
		if errors.As(err, &tgErr) {
			if strings.Contains(tgErr.Message, "Too Many Requests") && tgErr.RetryAfter != 0 {
				time.Sleep(time.Second * time.Duration(tgErr.RetryAfter))
				continue
			} else {
				log.Println("Error TG by Error: ", tgErr.Message, tgErr.RetryAfter)
			}
		} else if strings.Contains(err.Error(), tooMany) {
			valStr := strings.ReplaceAll(err.Error(), tooMany, "")
			sec, err := strconv.Atoi(strings.TrimSpace(valStr))
			if err == nil {
				time.Sleep(time.Second * time.Duration(sec))
				continue
			}
		}
		break
	}

	if err != nil {
		log.Println("Error TG Send: ", err)
	}
	return
}

func (s *Service) IsAuthorized(message *tgbotapi.Message) bool {

	for _, v := range s.mainGroups {
		if s.IsMemberMainGroup(message.From.ID, v) {
			return true
		}
	}

	return false
}

func (s *Service) IsMainGroup(id int64) bool {
	for _, v := range s.mainGroups {
		if v == id {
			return true
		}
	}
	return false
}

func (s *Service) IsMemberMainGroup(userID int, mgId int64) bool {
	member, err := s.bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: mgId,
		UserID: userID,
	})

	if err != nil {
		log.Println("Fail get member of main group", err)
		return false
	}

	if member.IsMember() || member.IsCreator() || member.IsAdministrator() {
		return true
	}

	return false
}

func (s *Service) mainLoop(updates tgbotapi.UpdatesChannel) {

	for update := range updates {

		if update.CallbackQuery != nil {
			if !s.IsAuthorized(update.CallbackQuery.Message) {
				continue
			}
			go s.HandleCallbackQuery(update)
			continue
		}

		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !s.IsAuthorized(update.Message) {
			continue
		}

		if update.Message.IsCommand() {
			go s.HandleCommand(update)
			continue
		}

		s.HandleTextMessage(update)
	}
	panic("Exit main loop")
}
