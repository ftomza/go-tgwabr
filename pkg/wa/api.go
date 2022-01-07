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

	"github.com/Rhymen/go-whatsapp/binary"

	"github.com/Rhymen/go-whatsapp"
	waproto "github.com/Rhymen/go-whatsapp/binary/proto"
)

func (s *Instance) ReadMessage(client, messageID string) (err error) {
	jid := s.PrepareClientJID(client)
	ch, err := s.conn.Read(jid, messageID)
	if err != nil {
		log.Println("WAInstance error Read message: ", err)
		return err
	}
	res := <-ch
	log.Println("WAInstance Read message result: ", res) //{"status":200}
	return nil
}

func (s *Instance) GetStatusLogin() bool {
	pong, err := s.conn.AdminTest()

	if !pong || err != nil {
		log.Println("WAInstance error pinging: ", err)
		return false
	}

	return true
}

func (s *Instance) GetStatusDevice() bool {
	if s.conn.Info != nil {
		return s.conn.Info.Connected
	}
	return false
}

func (s *Instance) GetStatusContacts() (bool, int, string) {
	desc := " not sync"
	if s.status.ContactsLoad.Desc != "" {
		desc = fmt.Sprintf(" sync: %s, %s", s.status.ContactsLoad.At.Format(time.RFC822), s.status.ContactsLoad.Desc)
	}
	if s.conn.Store != nil {
		return len(s.conn.Store.Contacts) > 0, len(s.conn.Store.Contacts), desc
	}
	return false, 0, desc
}

func (s *Instance) GetUnreadChat() (map[string]string, int, string) {
	res := map[string]string{}
	desc := " not sync"
	if s.status.ChatsLoad.Desc != "" {
		desc = fmt.Sprintf(" sync: %s, %s", s.status.ChatsLoad.At.Format(time.RFC822), s.status.ChatsLoad.Desc)
	}
	if s.conn.Store == nil {
		return res, 0, desc
	}
	for _, v := range s.conn.Store.Chats {
		if v.Unread != "0" {
			res[v.Jid] = v.Unread
		}
	}

	return res, len(s.conn.Store.Chats), desc
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
	if !ok {
		ch, err := s.conn.Exist(jid)
		if err == nil {
			status := <-ch
			log.Println("WAInstance Exist result: ", status)
			ok = strings.Contains(status, "\"status\":200,\"")
		} else {
			log.Println("WAInstance Exist error: ", err)
		}
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
	jid := s.PrepareClientJID(client)
	v, ok := s.conn.Store.Contacts[jid]
	if ok {
		if v.Name != "" {
			return v.Name
		}
		if v.Short != "" {
			return v.Short
		}
	}

	db, ok := appCtx.FromDB(s.ctx)
	if ok {
		items, err := db.GetContactsByWAClient(jid)
		if err != nil {
			log.Println("Get contact error: ", err)
		}
		if len(items) != 0 {
			itm := items[0]
			if itm.Name == "" {
				return itm.ShortName
			}
			return itm.Name
		}
	} else {
		log.Println("Store not ready")
	}

	if !pkg.StringInSlice(client, s.clients) {
		return "Not Sync"
	} else {
		return "New Client"
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
	var screenName string
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

	profilePicThumb, _ := s.conn.GetProfilePicThumb(client)
	profilePic := <-profilePicThumb

	thumbnail := ThumbUrl{}
	err = json.Unmarshal([]byte(profilePic), &thumbnail)

	if err != nil {
		return "", err
	}

	if thumbnail.Status == 404 {
		return "", nil
	}

	return thumbnail.EURL, nil

}

func (s *Instance) PrepareClientJID(client string) string {
	if strings.Count(client, "@") > 0 {
		return client
	}
	jidPrefix := "s.whatsapp.net"
	if len(strings.SplitN(client, "-", 2)) == 2 {
		jidPrefix = "g.us"
	} else if len(client) > 14 {
		jidPrefix = "g.us"
	}
	return fmt.Sprintf("%s@%s", client, jidPrefix)
}

func (s *Instance) PartsClientJID(jid string) (string, string) {
	if strings.Count(jid, "@") > 0 {
		parts := strings.Split(jid, "@")
		return parts[0], parts[1]
	}
	return "", ""
}

func (s *Instance) GetID() string {
	return fmt.Sprintf("%d", s.id)
}

func (s *Instance) SyncContacts() (bool, error) {
	node, err := s.conn.Contacts()
	if err != nil {
		log.Println("WAInstance SyncContact error: ", err)
		return false, err
	}

	if node == nil {
		log.Println("WAInstance SyncContact node is empty")
		return false, nil
	}

	if !(node.Description == "response" && node.Attributes["type"] == "contacts") {
		log.Println("WAInstance SyncContact node not type contacts, response type: ", node.Attributes["type"])
		return false, nil
	}

	c, ok := node.Content.([]interface{})
	if !ok {
		log.Printf("WAInstance SyncContact node content not Array %v\n", node.Content)
		return false, nil
	}

	var contactList []whatsapp.Contact
	for _, contact := range c {
		contactNode, ok := contact.(binary.Node)
		if !ok {
			continue
		}

		jid := strings.Replace(contactNode.Attributes["jid"], "@c.us", "@s.whatsapp.net", 1)
		s.conn.Store.Contacts[jid] = whatsapp.Contact{
			Jid:    jid,
			Notify: contactNode.Attributes["notify"],
			Name:   contactNode.Attributes["name"],
			Short:  contactNode.Attributes["short"],
		}
		contactList = append(contactList, s.conn.Store.Contacts[jid])
	}

	s.HandleContactList(contactList)

	return true, nil
}

func (s *Instance) SyncChats() (bool, error) {
	node, err := s.conn.Chats()
	if err != nil {
		log.Println("WAInstance SyncChats error: ", err)
		return false, err
	}

	if node == nil {
		log.Println("WAInstance SyncChats node is empty")
		return false, nil
	}

	if !(node.Description == "response" && node.Attributes["type"] == "chat") {
		log.Println("WAInstance SyncChats node not type chat, response type: ", node.Attributes["type"])
		return false, nil
	}

	c, ok := node.Content.([]interface{})
	if !ok {
		log.Printf("WAInstance SyncChats node content not Array %v\n", node.Content)
		return false, nil
	}
	var chatList []whatsapp.Chat
	for _, contact := range c {
		chatNode, ok := contact.(binary.Node)
		if !ok {
			continue
		}

		jid := strings.Replace(chatNode.Attributes["jid"], "@c.us", "@s.whatsapp.net", 1)
		s.conn.Store.Chats[jid] = whatsapp.Chat{
			Jid:             jid,
			Name:            chatNode.Attributes["name"],
			Unread:          chatNode.Attributes["count"],
			LastMessageTime: chatNode.Attributes["t"],
			IsMuted:         chatNode.Attributes["mute"],
			IsMarkedSpam:    chatNode.Attributes["spam"],
		}
		chatList = append(chatList, s.conn.Store.Chats[jid])
	}

	s.HandleChatList(chatList)

	return true, nil
}

func (s *Instance) sendStatusReady() {
	if !(strings.Count(s.status.ContactsLoad.Desc, "Load") > 0 &&
		strings.Count(s.status.ChatsLoad.Desc, "Load") > 0) {
		return
	}

	log.Println("WAInstance is sync")

	tg, ok := appCtx.FromTG(s.ctx)
	if !ok {
		return
	}

	_, _ = tg.SendMessage(s.id, "Bot is sync... check /status")
}
