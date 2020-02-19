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

func (s *Instance) GetStatusLogin() bool {
	pong, err := s.conn.AdminTest()

	if !pong || err != nil {
		log.Println("WAInstance error pinging: ", err)
		return false
	}

	return true
}

func (s *Instance) DoLogin() (ok bool, err error) {
	err = s.login(false)
	if err != nil {
		log.Println("WAInstance error login: ", err)
		return false, err
	}
	return true, nil
}

func (s *Instance) DoLogout() (bool, error) {
	err := s.conn.Logout()
	if err == nil {
		var session whatsapp.Session
		session, err = s.conn.Disconnect()
		if err == nil {
			err = s.writeSession(session)
		}
	}
	if err != nil {
		log.Println("WAInstance error logout: ", err)
		return false, err
	}
	return true, nil
}

func (s *Instance) ClientExist(client string) bool {
	jid := s.PrepareClientJID(client)
	_, ok := s.conn.Store.Contacts[jid]
	if !ok {
		ok = pkg.StringInSlice(jid, s.clients)
	}
	return ok
}

func (s *Instance) GetShortClient(client string) string {
	parts := strings.Split(client, "@")
	if len(parts) > 1 {
		return parts[0]
	}
	return client
}

func (s *Instance) GetClientName(client string) string {
	v, ok := s.conn.Store.Contacts[s.PrepareClientJID(client)]
	if ok {
		if v.Name == "" {
			return v.Short
		}
		return v.Name
	} else {
		if !pkg.StringInSlice(client, s.clients) {
			return "Not Exist"
		} else {
			return "New Client"
		}
	}
}

func (s *Instance) SendMessage(client, text, QuotedID, Quoted string) (msg *api.WAMessage, err error) {
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

func (s *Instance) SendImage(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
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

func (s *Instance) SendDocument(client string, reader io.Reader, mime string, fileName string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
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

func (s *Instance) SendAudio(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
	client = s.PrepareClientJID(client)
	item := whatsapp.AudioMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		Type:    mime,
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

func (s *Instance) SendVideo(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {
	client = s.PrepareClientJID(client)
	item := whatsapp.VideoMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		Type:    mime,
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

func (s *Instance) SendLocation(client string, lat, lon float64, QuotedID string, Quoted string) (msg *api.WAMessage, err error) {

	client = s.PrepareClientJID(client)
	item := whatsapp.LocationMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: client,
		},
		DegreesLongitude: lon,
		DegreesLatitude:  lat,
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
	s *Instance
}

func (h *historyHandler) ShouldCallSynchronously() bool {
	return true
}

func (h *historyHandler) HandleError(err error) {
	h.s.HandleError(err)
}

func (h *historyHandler) prepareText(message whatsapp.MessageInfo) string {
	screenName := "-"
	if message.FromMe {
		db, ok := appCtx.FromDB(h.s.ctx)
		screenName = "Me"
		if ok {
			mID := message.Id
			item, err := db.GetMessageByWA(mID)
			if err == nil && item != nil && item.TGUserName != "" {
				screenName = "@" + item.TGUserName
			}
		}
	} else {
		screenName = "Client"
	}

	date := time.Unix(int64(message.Timestamp), 0)
	return fmt.Sprintf("%s %s:", date, screenName)
}

func (h *historyHandler) HandleTextMessage(message whatsapp.TextMessage) {
	log.Println("ht: ", message.Info)
	message.Text = fmt.Sprintf("%s %s", h.prepareText(message.Info), message.Text)
	h.s.handleMessage(message, false)
}

func (h *historyHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	log.Println("ht: ", message.Info)
	message.Caption = fmt.Sprintf("%s %s", h.prepareText(message.Info), message.Caption)
	h.s.handleMessage(message, false)
}

func (h *historyHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	log.Println("ht: ", message.Info)
	message.Caption = fmt.Sprintf("%s %s", h.prepareText(message.Info), message.Caption)
	h.s.handleMessage(message, false)
}

func (h *historyHandler) HandleAudioMessage(message whatsapp.AudioMessage) {
	log.Println("ht: ", message.Info)
	h.s.handleMessage(message, false)
}

func (h *historyHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	log.Println("ht: ", message.Info)
	h.s.handleMessage(message, false)
}

func (h *historyHandler) HandleLocationMessage(message whatsapp.LocationMessage) {
	log.Println("ht: ", message.Info)
	h.s.handleMessage(message, false)
}

func (s *Instance) GetHistory(client string, size int) (err error) {
	client = s.PrepareClientJID(client)
	handler := &historyHandler{s: s}

	// load chat history and pass messages to the history handler to accumulate
	err = s.conn.LoadChatMessages(client, size, "", false, false, handler)
	if err != nil {
		return err
	}
	return nil
}

func (s *Instance) GetContactPhoto(client string) (result string, err error) {
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

func (s *Instance) PrepareClientJID(client string) string {
	if strings.Count(client, "@") > 0 {
		return client
	}
	jidPrefix := "s.whatsapp.net"
	if len(strings.SplitN(client, "-", 2)) == 2 {
		jidPrefix = "g.us"
	}
	return fmt.Sprintf("%s@%s", client, jidPrefix)
}

func (s *Instance) GetID() string {
	return fmt.Sprintf("%d", s.id)
}
