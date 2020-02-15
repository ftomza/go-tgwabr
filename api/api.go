package api

import "io"

var (
	DIRECTION_TG2WA = "tg2wa"
	DIRECTION_WA2TG = "wa2tg"
)

type WAMessage struct {
	Client       string
	Name         string
	MessageID    string
	Timestamp    uint64
	FwdMessageID string
}

type TGMessage struct {
	ChatID       int64
	UserName     string
	MessageID    int
	Timestamp    int
	FwdMessageID int
	FwdChatID    int64
}

type Message struct {
	WAID           string
	WAClient       string
	WAName         string
	WAMessageID    string
	WATimestamp    uint64
	WAFwdMessageID string
	TGChatID       int64
	TGUserName     string
	TGMessageID    int
	TGTimestamp    int
	TGFwdMessageID int
	TGFwdChatID    int64
	Direction      string
	Text           string
}

type Chat struct {
	WAID     string
	WAClient string
	TGChatID int64
}

type WA interface {
	GetStatusLogin() bool
	DoLogin() (bool, error)
	DoLogout() (bool, error)
	ClientExist(client string) bool
	GetClientName(client string) string
	SendMessage(client, text, QuotedID, Quoted string) (msg *WAMessage, err error)
	SendImage(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	SendDocument(client string, reader io.Reader, mime string, fileName string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	GetHistory(client string, size int) (result []string, err error)
	GetContactPhoto(client string) (result string, err error)
	GetShortClient(client string) string
	PrepareClientJID(client string) string
	GetID() string
}

type TG interface {
	SendQR(code string) (msg *TGMessage, err error)
	SendMessage(chatID int64, text string) (msg *TGMessage, err error)
	SendImage(chatID int64, reader io.Reader, caption string) (msg *TGMessage, err error)
	SendDocument(chatID int64, reader io.Reader, fileName string) (msg *TGMessage, err error)
	DeleteMessage(chatID int64, messageID int) (err error)
}

type Store interface {
	SaveMessage(message *Message) error
	GetMessageByWA(messageID string) (*Message, error)
	ExistMessageByWA(messageID string) bool
	ExistMessageByTG(messageID int, chatID int64) bool
	GetChatByClient(client string, id string) (*Chat, error)
	GetChatsByChatID(chatID int64) ([]*Chat, error)
	SaveChat(chat *Chat) error
	DeleteChat(chat *Chat) (bool, error)
}

type Cache interface {
	GetMembers() ([]int, error)
}
