package tg

import (
	"fmt"
	"io"
	"log"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/skip2/go-qrcode"
)

const chunkSize = 20

func (s *Service) SendQR(mgChatID int64, code string) (msg *api.TGMessage, err error) {

	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		log.Println(err)
	}
	qrFile := tgbotapi.FileBytes{
		Name:  "WhatsAppLoginQRCode",
		Bytes: png,
	}
	req := tgbotapi.NewPhotoUpload(mgChatID, qrFile)
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendMessage(chatID int64, text string) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewMessage(chatID, text)
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendImage(chatID int64, reader io.Reader, caption string) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileReader{
		Name:   caption,
		Reader: reader,
		Size:   -1,
	})
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendAudio(chatID int64, reader io.Reader) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewAudioUpload(chatID, tgbotapi.FileReader{
		Reader: reader,
		Size:   -1,
	})
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendVideo(chatID int64, reader io.Reader) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewVideoUpload(chatID, tgbotapi.FileReader{
		Reader: reader,
		Size:   -1,
	})
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendDocument(chatID int64, reader io.Reader, fileName string) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewDocumentUpload(chatID, tgbotapi.FileReader{
		Name:   fileName,
		Reader: reader,
		Size:   -1,
	})
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendLocation(chatID int64, lat, lon float64) (msg *api.TGMessage, err error) {
	req := tgbotapi.NewLocation(chatID, lat, lon)
	response, err := s.BotSend(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) DeleteMessage(chatID int64, messageID int) (err error) {
	_, err = s.bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	})
	return
}

func (s *Service) GetMembers() (members []int, err error) {
	return []int{}, nil
}

func (s *Service) UpdateStatMessage(chunk int) {

	db, ok := appCtx.FromDB(s.ctx)
	if !ok {
		log.Println("Error context db updateStatMessage")
		return
	}
	waSvc, ok := appCtx.FromWA(s.ctx)
	if !ok {
		log.Println("Module WhatsApp not ready")
		return
	}
	for _, v := range s.mainGroups {
		items, err := db.GetNotChatted(v, s.bot.Self.UserName)
		if err != nil {
			log.Println("Error get items updateStatMessage: ", err)
			return
		}
		wac, ok := waSvc.GetInstance(v)
		if !ok {
			log.Println("Instance WhatsApp not ready")
			continue
		}
		txt := ""
		for k, i := range items {
			if k < (chunk-1)*chunkSize {
				continue
			}
			if k > chunk*chunkSize {
				break
			}
			if i == nil {
				continue
			}

			if strings.Contains(i.WAClient, "@c.us") {
				continue
			}
			name := wac.GetClientName(i.WAClient)
			client := wac.GetShortClient(i.WAClient)

			//txt = fmt.Sprintf("%s\n <tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td></tr>", txt, name, client, i.TGUserName, i.Date.Format("2006-01-02"), i.Count)
			userName := ""
			if i.TGUserName != "" {
				userName = "@" + i.TGUserName
			}
			txt = fmt.Sprintf("%s\n - %s(%s) from [%s]>%s: %d, ur: %d", txt, name, client, userName, i.Date.Format("2006-01-02"), i.Count, i.CountUnread)
		}
		var row []tgbotapi.InlineKeyboardButton
		if chunk > 1 {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData("⬅ Prev", fmt.Sprintf("stat.get#%d", chunk-1)))
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("refresh 🔄 "+time.Now().Format("Mon 15:04"), "stat.refresh"))
		if len(items) > chunk*chunkSize {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData("Next ➡ ", fmt.Sprintf("stat.get#%d", chunk+1)))
		}

		grp, err := db.GetMainGroupByTGID(v)
		if err != nil {
			log.Println("Error get MainGroup updateStatMessage: ", err)
			return
		}

		if grp == nil {
			log.Println("Error get MainGroup not found: ", v)
			return
		}

		inlineKeyBoard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(row...))

		if grp.MessagePin > 0 && txt != "" {
			msg := tgbotapi.NewEditMessageText(v, grp.MessagePin, txt+"\n #pinstat")
			msg.ReplyMarkup = &inlineKeyBoard
			_, err = s.BotSend(msg)
			if err != nil && strings.Contains(err.Error(), "Bad Request: message to edit not found") {
				grp.MessagePin = -1
			}
		} else if grp.MessagePin > 0 && txt == "" {
			s.deletePin(grp, db)
		}
		if grp.MessagePin < 1 && txt != "" {
			msg := tgbotapi.NewMessage(v, txt+"\n #pinstat")
			msg.ReplyMarkup = &inlineKeyBoard
			resp, err := s.BotSend(msg)
			if err == nil {
				grp.MessagePin = resp.MessageID
				err = db.SaveMainGroup(grp)
			}
			if err == nil {
				_, err = s.bot.PinChatMessage(tgbotapi.PinChatMessageConfig{ChatID: v, MessageID: resp.MessageID, DisableNotification: true})
			}
			if err != nil {
				log.Printf("Error Send MSG on %d updateStatMessage: %s\n", v, err)
			}
		}
	}

}

func (s *Service) SendLog(text string) {
}

func (s *Service) GetMainGroups() []int64 {
	return s.mainGroups
}
