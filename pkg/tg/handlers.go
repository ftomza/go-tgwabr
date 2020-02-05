package tg

import (
	"fmt"
	"log"
	"net/http"
	"tgwabr/api"
	appCtx "tgwabr/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (s *Service) HandleTextMessage(update tgbotapi.Update) {

	if update.Message.Chat.ID == s.mainGroup {
		return
	}

	item := &api.Message{
		TGChatID:       update.Message.Chat.ID,
		TGUserName:     update.Message.From.UserName,
		TGMessageID:    update.Message.MessageID,
		TGTimestamp:    update.Message.Date,
		TGFwdMessageID: update.Message.ForwardFromMessageID,
		Direction:      api.DIRECTION_TG2WA,
	}
	if update.Message.ForwardFromChat != nil {
		item.TGFwdChatID = update.Message.ForwardFromChat.ID
	}

	if update.Message.Document != nil {
		item.Text = fmt.Sprintf("%s", update.Message.Document.FileName)
	} else if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
		item.Text = "PHOTO"
	} else {
		item.Text = update.Message.Text
	}

	if item.Text == "" {
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()

	db, ok := appCtx.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	wac, ok := appCtx.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	if db.ExistMessageByTG(item.TGMessageID, item.TGChatID) {
		return
	}

	chats, err := db.GetChatsByChatID(item.TGChatID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail send message, please send admin this error: %s", err)
		log.Println("Error get chats store: ", err)
		return
	}

	if len(chats) == 0 {
		msg.Text = "Chat not join!"
		return
	}
	var resp *api.WAMessage
	if update.Message.Document != nil {
		var urlFile string
		urlFile, err = s.bot.GetFileDirectURL(update.Message.Document.FileID)
		var respFile *http.Response
		if err == nil {
			respFile, err = http.Get(urlFile)
		}
		if err == nil {
			resp, err = wac.SendDocument(chats[0].WAClient, respFile.Body, update.Message.Document.MimeType, update.Message.Document.FileName, "", "")
		}
	} else if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
		v := (*update.Message.Photo)[len(*update.Message.Photo)-1]
		var urlFile string
		urlFile, err = s.bot.GetFileDirectURL(v.FileID)
		var respFile *http.Response
		if err == nil {
			respFile, err = http.Get(urlFile)
		}
		if err == nil {
			resp, err = wac.SendImage(chats[0].WAClient, respFile.Body, "image/jpeg", "", "")
		}
	} else {
		resp, err = wac.SendMessage(chats[0].WAClient, item.Text, "", "")
	}

	if err != nil {
		msg.Text = fmt.Sprintf("Fail send message, please send admin this error: %s", err)
		log.Println("Error Send message WA: ", err)
		return
	}

	item.WAClient = resp.Client
	item.WAMessageID = resp.MessageID
	item.WAName = resp.Name
	item.WAFwdMessageID = resp.FwdMessageID
	item.WATimestamp = resp.Timestamp

	err = db.SaveMessage(item)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail send message, please send admin this error: %s", err)
		log.Println("Error save Message store: ", err)
	}
}

func (s *Service) HandleCommand(update tgbotapi.Update) {
	switch update.Message.Command() {
	case "status":
		s.CommandStatus(update)
	case "login":
		s.CommandLogin(update)
	case "join":
		s.CommandJoin(update)
	case "leave":
		s.CommandLeave(update)
	case "history":
		s.CommandHistory(update)
	default:
		s.BotSend(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Command '%s' not implement", update.Message.Command())))
	}
}
