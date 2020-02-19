package tg

import (
	"fmt"
	"io"
	"log"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/skip2/go-qrcode"
)

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
	response, err := s.bot.Send(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendMessage(chatID int64, text string) (msg *api.TGMessage, err error) {

	req := tgbotapi.NewMessage(chatID, text)
	response, err := s.bot.Send(req)
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
	response, err := s.bot.Send(req)
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
	response, err := s.bot.Send(req)
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
	response, err := s.bot.Send(req)
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
	response, err := s.bot.Send(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendLocation(chatID int64, lat, lon float64) (msg *api.TGMessage, err error) {
	req := tgbotapi.NewLocation(chatID, lat, lon)
	response, err := s.bot.Send(req)
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

func (s *Service) UpdateStatMessage() {

	db, ok := appCtx.FromDB(s.ctx)
	if !ok {
		log.Println("Error context db updateStatMessage")
		return
	}

	for _, v := range s.mainGroups {
		items, err := db.GetNotChatted(v)
		if err != nil {
			log.Println("Error get items updateStatMessage: ", err)
			return
		}

		txt := ""
		for _, v := range items {
			if v == nil {
				continue
			}
			if strings.Contains(v.WAClient, "@c.us") || strings.Contains(v.WAClient, "@g.us") {
				continue
			}
			client := strings.ReplaceAll(v.WAClient, "@s.whatsapp.net", "")
			txt = fmt.Sprintf("%s\n - %s from %s: %d", txt, client, v.Date.Format("2006-01-02"), v.Count)
		}

		grp, err := db.GetMainGroupByTGID(v)
		if err != nil {
			log.Println("Error get MainGroup updateStatMessage: ", err)
			return
		}

		if grp.MessagePin > 0 && txt != "" {
			msg := tgbotapi.NewEditMessageText(v, grp.MessagePin, txt)
			_, err = s.bot.Send(msg)
			if err != nil && strings.Contains(err.Error(), "Bad Request: message to edit not found") {
				grp.MessagePin = -1
			}
		} else if grp.MessagePin > 0 && txt == "" {
			_, err = s.bot.UnpinChatMessage(tgbotapi.UnpinChatMessageConfig{ChatID: v})
			if err == nil {
				_, err = s.bot.DeleteMessage(tgbotapi.DeleteMessageConfig{
					ChatID:    v,
					MessageID: grp.MessagePin,
				})
			}
			if err == nil {
				grp.MessagePin = -1
				err = db.SaveMainGroup(grp)
			}
			if err != nil {
				log.Printf("Error delete MSG %d on %d updateStatMessage: %s\n", grp.MessagePin, v, err)
			}
		}
		if grp.MessagePin < 1 && txt != "" {
			msg := tgbotapi.NewMessage(v, txt)
			resp, err := s.bot.Send(msg)
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
