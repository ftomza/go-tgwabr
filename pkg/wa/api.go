package wa

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"tgwabr/pkg"
	"time"

	"github.com/Rhymen/go-whatsapp"
	waproto "github.com/Rhymen/go-whatsapp/binary/proto"
)

func (s *Service) GetStatusLogin() bool {
	pong, err := s.conn.AdminTest()

	if !pong || err != nil {
		log.Println("WA error pinging: ", err)
		return false
	}

	return true
}

func (s *Service) DoLogin() (ok bool, err error) {
	err = s.login(false)
	if err != nil {
		log.Println("WA error login: ", err)
		return false, err
	}
	return true, nil
}

func (s *Service) ClientExist(client string) bool {
	jid := s.PrepareClientJID(client)
	_, ok := s.conn.Store.Contacts[jid]
	if !ok {
		ok = pkg.StringInSlice(client, s.clients)
	}
	return ok
}

func (s *Service) GetShortClient(client string) string {
	parts := strings.Split(client, "@")
	if len(parts) > 1 {
		return parts[0]
	}
	return client
}

func (s *Service) GetClientName(client string) string {
	v, ok := s.conn.Store.Contacts[s.PrepareClientJID(client)]
	if ok {
		return v.Name
	} else {
		if !pkg.StringInSlice(client, s.clients) {
			return "Not Exist"
		} else {
			return "New Client"
		}
	}
}

func (s *Service) SendMessage(client, text, QuotedID, Quoted string) (msg *api.WAMessage, err error) {
	client = s.PrepareClientJID(client)
	item := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		Text: text,
	}
	if len(QuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &Quoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: QuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     client,
		}

		item.ContextInfo = ctxQuotedInfo
	}

	msgId, err := s.conn.Send(item)
	if err != nil {
		return nil, err
	}
	name := s.GetClientName(client)
	return &api.WAMessage{
		Client:    client,
		Name:      name,
		MessageID: msgId,
		Timestamp: uint64(time.Now().Unix()),
	}, nil
}

func (s *Service) SendImage(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
	client = s.PrepareClientJID(client)
	item := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		Type: mime,

		Content: reader,
	}
	if len(QuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &Quoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: QuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     client,
		}

		item.ContextInfo = ctxQuotedInfo
	}

	msgId, err := s.conn.Send(item)
	if err != nil {
		return nil, err
	}
	name := s.GetClientName(client)
	return &api.WAMessage{
		Client:    client,
		Name:      name,
		MessageID: msgId,
		Timestamp: uint64(time.Now().Unix()),
	}, nil
}

func (s *Service) SendDocument(client string, reader io.Reader, mime string, fileName string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
	client = s.PrepareClientJID(client)
	item := whatsapp.DocumentMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		Type:     mime,
		FileName: fileName,
		Content:  reader,
	}
	if len(QuotedID) != 0 {
		msgQuotedProto := waproto.Message{
			Conversation: &Quoted,
		}

		ctxQuotedInfo := whatsapp.ContextInfo{
			QuotedMessageID: QuotedID,
			QuotedMessage:   &msgQuotedProto,
			Participant:     client,
		}

		item.ContextInfo = ctxQuotedInfo
	}

	msgId, err := s.conn.Send(item)
	if err != nil {
		return nil, err
	}
	name := s.GetClientName(client)
	return &api.WAMessage{
		Client:    client,
		Name:      name,
		MessageID: msgId,
		Timestamp: uint64(time.Now().Unix()),
	}, nil
}

type historyHandler struct {
	s        *Service
	messages []string
}

func (h *historyHandler) ShouldCallSynchronously() bool {
	return true
}

func (h *historyHandler) HandleError(err error) {
	h.messages = append(h.messages, fmt.Sprintf("Fail get History chat, please send admin this error: %s", err))
}

func (h *historyHandler) HandleTextMessage(message whatsapp.TextMessage) {
	screenName := "-"
	if message.Info.FromMe {
		db, ok := appCtx.FromDB(h.s.ctx)
		screenName = "Me"
		if ok {
			mID := message.Info.Id
			item, err := db.GetMessageByWA(mID)
			if err == nil && item != nil && item.TGUserName != "" {
				screenName = "@" + item.TGUserName
			}
		}
	} else {
		screenName = "Client"
	}

	date := time.Unix(int64(message.Info.Timestamp), 0)
	h.messages = append(h.messages, fmt.Sprintf("%s %s: %s", date,
		screenName, message.Text))

}

func (s *Service) GetHistory(client string, size int) (result []string, err error) {
	client = s.PrepareClientJID(client)
	handler := &historyHandler{s: s}

	// load chat history and pass messages to the history handler to accumulate
	err = s.conn.LoadChatMessages(client, size, "", false, false, handler)
	if err != nil {
		return nil, err
	}
	return handler.messages, nil
}

func (s *Service) GetContactPhoto(client string) (result string, err error) {
	client = s.PrepareClientJID(client)
	type ThumbUrl struct {
		EURL   string `json:"eurl"`
		Tag    string `json:"tag"`
		Status int64  `json:"status"`
	}

	// RemoteJID is the sender ID the one who send the text or the media message or basically the WhatsappID just pass it here
	profilePicThumb, _ := s.conn.GetProfilePicThumb(client)
	profilePic := <-profilePicThumb

	thumbnail := ThumbUrl{}
	err = json.Unmarshal([]byte(profilePic), &thumbnail)

	if err != nil {
		// print error
	}

	if thumbnail.Status == 404 {
		// meaning thumbnail is not available because the person has no profile pic so what i did is return empty string
	}

	// Basically the EURL is what holds the profile picture of the person
	return thumbnail.EURL, nil

}

func (s *Service) PrepareClientJID(client string) string {
	if strings.Count(client, "@") > 0 {
		return client
	}
	jidPrefix := "s.whatsapp.net"
	if len(strings.SplitN(client, "-", 2)) == 2 {
		jidPrefix = "g.us"
	}
	return fmt.Sprintf("%s@%s", client, jidPrefix)
}
