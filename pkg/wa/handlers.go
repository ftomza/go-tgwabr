package wa

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"tgwabr/pkg"
	"time"

	"github.com/cristalinojr/go-whatsapp"
)

func (s *Instance) handleMessage(message interface{}, doSave bool) {
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
	case whatsapp.AudioMessage:
		info = m.Info
		msg.Text = m.Type
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	case whatsapp.VideoMessage:
		info = m.Info
		msg.Text = fmt.Sprintf("%s (%s)", m.Caption, m.Type)
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	case whatsapp.LocationMessage:
		info = m.Info
		msg.Text = m.Name
		msg.WAFwdMessageID = m.ContextInfo.QuotedMessageID
	default:
		log.Println(fmt.Sprintf("Type not implement %T", m))
		return
	}

	if info.Timestamp < s.pointTime && doSave {
		return
	}

	if info.RemoteJid == "status@broadcast" {
		return
	}

	name := s.conn.Store.Contacts[info.RemoteJid].Name
	fromName := s.conn.Store.Contacts[info.SenderJid].Name
	msg = &api.Message{
		MGID:           s.GetID(),
		WAClient:       info.RemoteJid,
		WAName:         name,
		WAFromName:     fromName,
		WAFromClient:   info.SenderJid,
		WAMessageID:    info.Id,
		WATimestamp:    info.Timestamp,
		WAFwdMessageID: msg.WAFwdMessageID,
		Chatted:        api.ChattedNo,
		MessageStatus:  int(info.Status),
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

	if doSave {
		if db.ExistMessageByWA(msg.WAMessageID) {
			return
		}

		err := db.SaveMessage(msg)
		if err != nil {
			log.Println("Save store error: ", err)
		}
	}

	tg, ok := appCtx.FromTG(s.ctx)
	if !ok {
		fmt.Println(msg)
		return
	}

	chat, err := db.GetChatByClient(info.RemoteJid, s.GetID())
	if err != nil {
		log.Println("Get chat store error: ", err)
	}
	chatID := s.id
	if chat != nil {
		chatID = chat.TGChatID
		msg.Chatted = api.ChattedYes
		msg.TGUserName = chat.TGUserName
		msg.Session = chat.Session
	} else if info.FromMe {
		return
	}
	tgMsg := &api.TGMessage{}
	switch m := message.(type) {
	case whatsapp.TextMessage:
		txt := msg.Text
		if chat == nil {
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
			if _, err = s.conn.LoadMediaInfo(m.Info.RemoteJid, m.Info.Id, m.Info.FromMe); err == nil {
				raw, err = m.Download()
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
	case whatsapp.AudioMessage:
		var raw []byte
		raw, err = m.Download()
		if err == nil {
			buf := bytes.NewReader(raw)
			tgMsg, err = tg.SendAudio(chatID, buf)
		}
	case whatsapp.VideoMessage:
		var raw []byte
		raw, err = m.Download()
		if err == nil {
			buf := bytes.NewReader(raw)
			tgMsg, err = tg.SendVideo(chatID, buf)
		}
	case whatsapp.LocationMessage:
		tgMsg, err = tg.SendLocation(chatID, m.DegreesLatitude, m.DegreesLongitude)
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
	msg.TGFwdMessageID = tgMsg.FwdMessageID
	msg.Direction = api.DirectionWa2tg
	if msg.TGUserName == "" {
		msg.TGUserName = tgMsg.UserName
	}

	if doSave {
		err = s.ReadMessage(msg.WAClient, msg.WAMessageID)
		if err != nil {
			log.Println("Error read message: ", err)
		}

		err = db.SaveMessage(msg)
		if err != nil {
			log.Println("Save store error: ", err)
		}
		tg.UpdateStatMessage(1)
	}
}

func (s *Instance) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok || errors.Is(err, whatsapp.ErrInvalidWsData) {
		errp := err
		if e != nil {
			errp = e.Err
		}
		log.Printf("Connection failed, underlying error: %v", errp)
		log.Println("WAInstance Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("WAInstance Reconnecting...")
		err = s.conn.Restore()
		if err != nil {
			log.Println("Restore failed WAInstance: ", err)
		}
	} else {
		log.Println("error WAInstance occoured: ", err)
	}
}

func (s *Instance) HandleTextMessage(message whatsapp.TextMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleImageMessage(message whatsapp.ImageMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleVideoMessage(message whatsapp.VideoMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleAudioMessage(message whatsapp.AudioMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleLocationMessage(message whatsapp.LocationMessage) {
	s.handleMessage(message, true)
}

func (s *Instance) HandleJsonMessage(message string) {
	log.Println("WAInstance message JSON: ", message)
}

func (s *Instance) HandleContactMessage(_ whatsapp.ContactMessage) {
}

func (s *Instance) HandleContactList(contacts []whatsapp.Contact) {
	log.Printf("WAInstance contacts load: %d\n", len(contacts))
	s.status.ContactsLoad.At = time.Now()
	if len(contacts) > 0 {
		s.status.ContactsLoad.Desc = fmt.Sprintf("Load: %d", len(contacts))
		s.sendStatusReady()
		db, ok := appCtx.FromDB(s.ctx)
		if !ok {
			log.Println("Store not ready")
			return
		}
		for _, v := range contacts {
			phone, tp := s.PartsClientJID(v.Jid)
			if tp == "g.us" {
				continue
			}
			if v.Name == "" {
				continue
			}
			err := db.SaveContact(&api.Contact{
				Phone:     phone,
				WAClient:  v.Jid,
				Name:      v.Name,
				ShortName: v.Short,
			})
			if err != nil {
				log.Println("Sync Store contacts error: ", err)
			}
		}
	} else {
		s.status.ContactsLoad.Desc = "Receive empty"
	}
}

func (s *Instance) HandleChatList(chats []whatsapp.Chat) {
	log.Printf("WAInstance chat load: %d\n", len(chats))
	s.status.ChatsLoad.At = time.Now()
	if len(chats) > 0 {
		s.status.ChatsLoad.Desc = fmt.Sprintf("Load: %d", len(chats))
		s.sendStatusReady()
	} else {
		s.status.ChatsLoad.Desc = "Receive empty"
	}
}
