package wa

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"tgwabr/pkg"
	"time"

	"github.com/Rhymen/go-whatsapp"
)

func (s *Service) handleMessage(message interface{}) {
	msg := &api.Message{}
	var info whatsapp.MessageInfo
	switch m := message.(type) {
	case whatsapp.TextMessage:
		info = m.Info
		msg.Text = m.Text
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	case whatsapp.ImageMessage:
		info = m.Info
		msg.Text = fmt.Sprintf("%s (%s)", m.Caption, m.Type)
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	case whatsapp.DocumentMessage:
		info = m.Info
		msg.Text = fmt.Sprintf("%s (%s)", m.Title, m.Type)
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	default:
		log.Println(fmt.Sprintf("Type not implement %T", m))
		return
	}

	if info.Timestamp < s.pointTime {
		return
	}

	if info.RemoteJid == "status@broadcast" {
		return
	}

	name := s.conn.Store.Contacts[info.RemoteJid].Name
	msg = &api.Message{
		WAClient:       info.RemoteJid,
		WAName:         name,
		WAMessageID:    info.Id,
		WATimestamp:    info.Timestamp,
		WAFwdMessageID: msg.WAFwdMessageID,
		Text:           msg.Text,
	}

	if info.FromMe {
		msg.WAName = "Self"
		msg.WAClient = s.conn.Info.Wid
	}

	if !pkg.StringInSlice(msg.WAClient, s.clients) {
		s.clients = append(s.clients, msg.WAClient)
	}

	db, ok := appCtx.FromDB(s.ctx)
	if !ok {
		log.Println("Store not ready")
		return
	}

	if db.ExistMessageByWA(msg.WAMessageID) {
		return
	}

	err := db.SaveMessage(msg)
	if err != nil {
		log.Println("Save store error: ", err)
	}

	tg, ok := appCtx.FromTG(s.ctx)
	if !ok {
		fmt.Println(msg)
		return
	}

	chat, err := db.GetChatByClient(info.RemoteJid)
	if err != nil {
		log.Println("Get chat store error: ", err)
	}
	chatID := int64(0)
	if chat != nil {
		chatID = chat.TGChatID
	} else if info.FromMe {
		return
	}
	tgMsg := &api.TGMessage{}
	switch m := message.(type) {
	case whatsapp.TextMessage:
		txt := msg.Text
		if chatID == 0 {
			builder := strings.Builder{}
			builder.WriteString(fmt.Sprintf("Client %s(%s):\n", msg.WAName, s.GetShortClient(msg.WAClient)))
			builder.WriteString(txt)
			txt = builder.String()
		}
		tgMsg, err = tg.SendMessage(chatID, txt)
	case whatsapp.ImageMessage:
		var raw []byte
		raw, err = m.Download()
		if err != nil {
			if err != whatsapp.ErrMediaDownloadFailedWith410 && err != whatsapp.ErrMediaDownloadFailedWith404 {
				if _, err = s.conn.LoadMediaInfo(m.Info.RemoteJid, m.Info.Id, strconv.FormatBool(m.Info.FromMe)); err == nil {
					raw, err = m.Download()
				}
			}
		}
		if err == nil {
			buf := bytes.NewReader(raw)
			tgMsg, err = tg.SendImage(chatID, buf, m.Caption)
		}
	case whatsapp.DocumentMessage:
		var raw []byte
		raw, err = m.Download()
		if err == nil {
			buf := bytes.NewReader(raw)
			tgMsg, err = tg.SendDocument(chatID, buf, m.FileName)
		}
	default:
		return
	}

	if err != nil {
		log.Println("Send message tg error: ", err)
		return
	}

	msg.TGChatID = tgMsg.ChatID
	msg.TGMessageID = tgMsg.MessageID
	msg.TGTimestamp = tgMsg.Timestamp
	msg.TGUserName = tgMsg.UserName
	msg.TGFwdMessageID = tgMsg.FwdMessageID
	msg.Direction = api.DIRECTION_WA2TG

	err = db.SaveMessage(msg)
	if err != nil {
		log.Println("Save store error: ", err)
	}
}

func (s *Service) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok || errors.Is(err, whatsapp.ErrInvalidWsData) {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("WA Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("WA Reconnecting...")
		err = s.conn.Restore()
		if err != nil {
			log.Println("Restore failed WA: ", err)
		}
	} else {
		log.Println("error WA occoured: ", err)
	}
}

func (s *Service) HandleTextMessage(message whatsapp.TextMessage) {
	s.handleMessage(message)
}

func (s *Service) HandleImageMessage(message whatsapp.ImageMessage) {
	s.handleMessage(message)
}

func (s *Service) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	s.handleMessage(message)
}

func (s *Service) HandleJsonMessage(_ string) {
}

func (s *Service) HandleContactMessage(_ whatsapp.ContactMessage) {
}
