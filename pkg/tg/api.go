package tg

import (
	"io"
	"log"
	"tgwabr/api"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/skip2/go-qrcode"
)

func (s *Service) SendQR(code string) (msg *api.TGMessage, err error) {

	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		log.Println(err)
	}
	qrFile := tgbotapi.FileBytes{
		Name:  "WhatsAppLoginQRCode",
		Bytes: png,
	}
	req := tgbotapi.NewPhotoUpload(s.mainGroup, qrFile)
	response, err := s.bot.Send(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendMessage(chatID int64, text string) (msg *api.TGMessage, err error) {
	if chatID == 0 {
		chatID = s.mainGroup
	}
	req := tgbotapi.NewMessage(chatID, text)
	response, err := s.bot.Send(req)
	if err != nil {
		return nil, err
	}
	return Message(response).ToAPIMessage(), nil
}

func (s *Service) SendImage(chatID int64, reader io.Reader, caption string) (msg *api.TGMessage, err error) {
	if chatID == 0 {
		chatID = s.mainGroup
	}
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

func (s *Service) SendDocument(chatID int64, reader io.Reader, fileName string) (msg *api.TGMessage, err error) {
	if chatID == 0 {
		chatID = s.mainGroup
	}
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
