package tg

import (
	"io"
	"log"
	"tgwabr/api"

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
