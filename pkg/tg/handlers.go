package tg

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (s *Service) HandleTextMessage(update tgbotapi.Update) {

	chatID := update.Message.Chat.ID

	if s.IsMainGroup(chatID) {
		return
	}

	item := &api.Message{
		TGChatID:       update.Message.Chat.ID,
		TGUserName:     update.Message.From.UserName,
		TGMessageID:    update.Message.MessageID,
		TGTimestamp:    update.Message.Date,
		TGFwdMessageID: update.Message.ForwardFromMessageID,
		Chatted:        api.ChattedYes,
		Direction:      api.DirectionTg2wa,
	}
	if update.Message.ForwardFromChat != nil {
		item.TGFwdChatID = update.Message.ForwardFromChat.ID
	}

	if update.Message.Document != nil {
		item.Text = update.Message.Document.FileName
	} else if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
		item.Text = "PHOTO"
	} else if update.Message.Audio != nil {
		item.Text = fmt.Sprintf("AUDIO %s", update.Message.Audio.Title)
	} else if update.Message.Video != nil {
		item.Text = fmt.Sprintf("VIDEO %s", update.Message.Video.MimeType)
	} else if update.Message.Location != nil {
		item.Text = fmt.Sprintf("LOCATION")
	} else {
		item.Text = update.Message.Text
	}

	if item.Text == "" {
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	db, ok := appCtx.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	waSvc, ok := appCtx.FromWA(s.ctx)
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

	chat := chats[0]
	mgChatID, _ := strconv.ParseInt(chat.MGID, 10, 64)
	wac, ok := waSvc.GetInstance(mgChatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	var resp *api.WAMessage
	var respFile *http.Response
	if update.Message.Audio != nil {
		err, respFile = s.getFileResponse(update.Message.Audio.FileID)
		if err == nil {
			resp, err = wac.SendAudio(chat.WAClient, respFile.Body, update.Message.Audio.MimeType, "", "")
		}
	} else if update.Message.Video != nil {
		err, respFile = s.getFileResponse(update.Message.Video.FileID)
		if err == nil {
			resp, err = wac.SendAudio(chat.WAClient, respFile.Body, update.Message.Video.MimeType, "", "")
		}
	} else if update.Message.Location != nil {
		resp, err = wac.SendLocation(chat.WAClient, update.Message.Location.Latitude, update.Message.Location.Longitude, "", "")
	} else if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
		v := (*update.Message.Photo)[len(*update.Message.Photo)-1]
		err, respFile = s.getFileResponse(v.FileID)
		if err == nil {
			resp, err = wac.SendImage(chat.WAClient, respFile.Body, "image/jpeg", "", "")
		}
	} else if update.Message.Document != nil {
		err, respFile = s.getFileResponse(update.Message.Document.FileID)
		if err == nil {
			resp, err = wac.SendDocument(chat.WAClient, respFile.Body, update.Message.Document.MimeType, update.Message.Document.FileName, "", "")
		}
	} else {
		resp, err = wac.SendMessage(chat.WAClient, item.Text, "", "")
	}

	if err != nil {
		msg.Text = fmt.Sprintf("Fail send message, please send admin this error: %s", err)
		log.Println("Error Send message WAInstance: ", err)
		return
	}

	item.MGID = wac.GetID()
	item.WAClient = resp.Client
	item.WAMessageID = resp.MessageID
	item.WAName = resp.Name
	item.WAFwdMessageID = resp.FwdMessageID
	item.WATimestamp = resp.Timestamp
	item.Session = chat.Session

	err = db.SaveMessage(item)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail send message, please send admin this error: %s", err)
		log.Println("Error save Message store: ", err)
	}
}

func (s *Service) getFileResponse(fileID string) (err error, respFile *http.Response) {
	var urlFile string
	urlFile, err = s.bot.GetFileDirectURL(fileID)
	if err == nil {
		respFile, err = http.Get(urlFile)
	}
	return err, respFile
}

func (s *Service) HandleCommand(update tgbotapi.Update) {
	switch update.Message.Command() {
	case "status":
		s.CommandStatus(update)
	case "set":
		s.CommandSet(update)
	case "login":
		s.CommandLogin(update)
	case "logout":
		s.CommandLogout(update)
	case "join":
		s.CommandJoin(update, "", "")
	case "leave":
		s.CommandLeave(update)
	case "history":
		s.CommandHistory(update)
	case "stat":
		s.CommandStat(update)
	case "check_client":
		s.CommandCheckClient(update)
	case "alias":
		s.CommandAlias(update)
	case "set_logger":
		s.CommandSetLogger(update)
	case "sync":
		s.CommandSync(update)
	case "contact":
		s.CommandContact(update)
	case "repined":
		s.CommandRePined(update)
	case "restart":
		s.CommandRestart(update)
	case "autoreplay":
		s.CommandAutoReplay(update)
	case "somethingelse":
		s.CommandSomethingElse(update, "", "")
	default:
		_, _ = s.BotSend(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Command '%s' not implement", update.Message.Command())))
	}
}

func (s *Service) HandleCallbackQuery(update tgbotapi.Update) {
	parts := strings.SplitN(update.CallbackQuery.Data, ".", 2)
	switch parts[0] {
	case "stat":
		s.CallbackQueryStat(update.CallbackQuery, parts)
	case "somethingelse":
		s.CallbackQuerySomethingElse(update.CallbackQuery, parts)
	case "chat":
		s.CallbackQueryChat(update, parts)
	default:
		_, _ = s.BotSend(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("Callback data '%s' not implement", parts[0])))
	}
}
