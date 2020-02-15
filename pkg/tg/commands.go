package tg

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tgwabr/api"
	"tgwabr/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (s *Service) CommandStatus(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(s.mainGroup, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()

	wa, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	if update.Message.Chat.ID != s.mainGroup {
		msg.ChatID = update.Message.Chat.ID
		msg.Text = "Command work only 'Main group'"
		return
	}

	if wa.GetStatusLogin() {
		msg.Text = "Online"
	} else {
		msg.Text = "Offline"
	}
}

func (s *Service) CommandHistory(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()

	wac, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	items, err := db.GetChatsByChatID(update.Message.Chat.ID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get History chat, please send admin this error: %s", err)
		log.Println("Error save chat store: ", err)
		return
	}

	if len(items) == 0 {
		msg.Text = fmt.Sprintf("Chat not joined!")
		return
	}

	client := items[0].WAClient
	name := wac.GetClientName(client)

	params := update.Message.CommandArguments()
	params = strings.ToLower(strings.TrimSpace(params))
	size := 0
	if size, err = strconv.Atoi(params); err != nil {
		size = 10
	}

	messages, err := wac.GetHistory(client, size)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get History chat for '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error get History: ", err)
	}

	for _, v := range messages {
		msg.Text = v
		s.BotSend(msg)
	}
	msg.Text = ""
}

func (s *Service) CommandLogin(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(s.mainGroup, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()

	wa, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	if update.Message.Chat.ID != s.mainGroup {
		msg.ChatID = update.Message.Chat.ID
		msg.Text = "Command work only 'Main group'"
		return
	}

	ok, err := wa.DoLogin()
	if err != nil {
		msg.Text = fmt.Sprintf("Error login: %s", err)
	} else if ok {
		msg.Text = "Login OK"
	} else {
		msg.Text = "Login FAIL"
	}
}

func (s *Service) CommandLogout(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(s.mainGroup, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()

	wa, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	if update.Message.Chat.ID != s.mainGroup {
		msg.ChatID = update.Message.Chat.ID
		msg.Text = "Command work only 'Main group'"
		return
	}

	ok, err := wa.DoLogout()
	if err != nil {
		msg.Text = fmt.Sprintf("Error login: %s", err)
	} else if ok {
		msg.Text = "Logout OK"
	} else {
		msg.Text = "Logout FAIL"
	}
}

func (s *Service) CommandJoin(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()
	if update.Message.Chat.ID == s.mainGroup {
		msg.Text = "Main group not join client"
		return
	}
	wac, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}
	client := update.Message.CommandArguments()
	client = strings.ToLower(strings.TrimSpace(client))
	if client == "" {
		msg.Text = "Client not set"
		return
	}
	if client == "all" {
		msg.Text = "ALL not work :'("
		return
	}
	if client != "check" && !wac.ClientExist(client) {
		msg.Text = fmt.Sprintf("Client '%s' not found", client)
		return
	}

	name := wac.GetClientName(client)

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	chat := api.Chat{
		WAID:     wac.GetID(),
		WAClient: wac.PrepareClientJID(client),
		TGChatID: update.Message.Chat.ID,
	}

	items, err := db.GetChatsByChatID(chat.TGChatID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail join chat '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error save chat store: ", err)
		return
	}

	if len(items) > 0 {
		name := wac.GetClientName(items[0].WAClient)
		msg.Text = fmt.Sprintf("Chat already joined to client '%s(%s)'", name, items[0].WAClient)
		return
	}

	err = db.SaveChat(&chat)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail join chat '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error save chat store: ", err)
		return
	}

	msgJoin := tgbotapi.NewMessage(s.mainGroup, fmt.Sprintf("Chat %s(%s) join to @%s", name, client, update.Message.From.UserName))
	s.BotSend(msgJoin)

	_, _ = s.bot.SetChatTitle(tgbotapi.SetChatTitleConfig{
		ChatID: chat.TGChatID,
		Title:  fmt.Sprintf("Chat with %s(%s)", name, client),
	})

	raw, err := wac.GetContactPhoto(client)
	if raw != "" {
		resp, err := s.bot.SetChatPhoto(tgbotapi.SetChatPhotoConfig{
			BaseFile: tgbotapi.BaseFile{
				BaseChat: tgbotapi.BaseChat{
					ChatID: chat.TGChatID,
				},
				File: tgbotapi.FileBytes{
					Bytes: getPhotoByte(raw),
				},
			},
		})
		if err != nil {
			log.Println(err)
		}
		log.Println(resp)
	}

	msg.Text = fmt.Sprintf("Join '%s(%s)' OK", name, client)
}

func (s *Service) CommandLeave(update tgbotapi.Update) {
	var err error
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	defer func() {
		if msg.Text != "" {
			s.BotSend(msg)
		}
	}()
	wac, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	client := update.Message.CommandArguments()
	client = strings.ToLower(strings.TrimSpace(client))

	chats, err := db.GetChatsByChatID(update.Message.Chat.ID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail leave 'all' chat, please send admin this error: %s", err)
		log.Println("Error get chats store: ", err)
		return
	}
	txt := "Leave chats: \n"
	for _, v := range chats {
		name := wac.GetClientName(v.WAClient)
		ok, err = db.DeleteChat(v)
		if err != nil {
			msg.Text = fmt.Sprintf("Fail leave '%s(%s)' chat, please send admin this error: %s", name, v.WAClient, err)
			log.Println("Error delete chat store: ", err)
			return
		}
		txt = txt + fmt.Sprintf(" - '%s(%s)' OK\n", name, v.WAClient)
		msgJoin := tgbotapi.NewMessage(s.mainGroup, fmt.Sprintf("@%s leave chat %s(%s)", update.Message.From.UserName, name, wac.GetShortClient(v.WAClient)))
		s.BotSend(msgJoin)
	}
	msg.Text = txt

	_, _ = s.bot.SetChatTitle(tgbotapi.SetChatTitleConfig{
		ChatID: update.Message.Chat.ID,
		Title:  fmt.Sprintf("H.W.Bot Free chat"),
	})

	_, _ = s.bot.DeleteChatPhoto(tgbotapi.DeleteChatPhotoConfig{ChatID: update.Message.Chat.ID})
}

func getPhotoByte(path string) []byte {
	resp, err := http.Get(path)
	if err != nil {
		log.Println(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return b
}
