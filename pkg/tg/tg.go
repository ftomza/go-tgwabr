package tg

import (
	"context"
	"log"
	"os"
	"strconv"
	"tgwabr/api"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Service struct {
	ctx       context.Context
	bot       *tgbotapi.BotAPI
	mainGroup int64

	api.TG
}

func New(ctx context.Context) (service *Service, err error) {

	service = &Service{ctx: ctx}

	mainGroupStr := os.Getenv("TG_MAIN_GROUP")
	mainGroup, err := strconv.ParseInt(mainGroupStr, 10, 64)
	service.mainGroup = mainGroup
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

func (s *Service) BotSend(msg tgbotapi.Chattable) (response tgbotapi.Message) {
	var err error
	response, err = s.bot.Send(msg)
	if err != nil {
		log.Println("Error TG Send: ", err)
	}
	return
}

func (s *Service) IsAuthorized(update tgbotapi.Update) bool {

	member, err := s.bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: s.mainGroup,
		UserID: update.Message.From.ID,
	})

	if err != nil {
		log.Println("Fail get member of main group", err)
		return false
	}

	if !(member.IsMember() || member.IsCreator() || member.IsAdministrator()) {
		return false
	}

	return true
}

func (s *Service) mainLoop(updates tgbotapi.UpdatesChannel) {

	for update := range updates {

		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !s.IsAuthorized(update) {
			continue
		}

		if update.Message.IsCommand() {
			s.HandleCommand(update)
			continue
		}

		s.HandleTextMessage(update)
	}
}
