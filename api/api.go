package api

import (
	"io"
	"time"
)

var (
	DirectionTg2wa = "tg2wa"
	DirectionWa2tg = "wa2tg"
	ChattedYes     = "yes"
	ChattedNo      = "no"
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
	MGID           string
	WAClient       string
	WAName         string
	WAFromClient   string
	WAFromName     string
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
	Chatted        string
	Answered       uint64
	MessageStatus  int
	Text           string
	Session        string
}

type Chat struct {
	MGID       string
	WAClient   string
	TGChatID   int64
	TGUserName string
	Session    string
}

type MainGroup struct {
	TGChatID     int64
	Name         string
	MessagePin   int
	LoggerChatID int64
}

type Stat struct {
	Date       time.Time
	TGUserName string
	WAName     string
	WAClient   string
	Session    *time.Time
	Answered   *float64
	CountIn    int
	CountOut   int
}

type StatDay struct {
	Date        time.Time
	WAClient    string
	TGUserName  string
	Count       int
	CountUnread int
}

type Alias struct {
	MGID     string
	WAClient string
	Name     string
}

type Contact struct {
	Phone     string
	Email     string
	WAClient  string
	TGUserID  int
	Name      string
	ShortName string
}

type WA interface {
	GetInstance(id int64) (WAInstance, bool)
}

type WAInstance interface {
	ReadMessage(client, messageID string) (err error)
	GetUnreadChat() (map[string]string, int, string)
	GetStatusContacts() (bool, int, string)
	GetStatusDevice() bool
	GetStatusLogin() bool
	DoLogin() (bool, error)
	DoLogout() (bool, error)
	ClientExist(client string) bool
	GetClientName(client string) string
	SendMessage(client, text, QuotedID, Quoted string) (msg *WAMessage, err error)
	SendImage(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	SendDocument(client string, reader io.Reader, mime string, fileName string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	SendAudio(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	SendVideo(client string, reader io.Reader, mime string, QuotedID string, Quoted string) (msg *WAMessage, err error)
	SendLocation(client string, lat, lon float64, QuotedID string, Quoted string) (msg *WAMessage, err error)
	GetHistory(client string, size int) (err error)
	GetContactPhoto(client string) (result string, err error)
	GetShortClient(client string) string
	PrepareClientJID(client string) string
	PartsClientJID(jid string) (string, string)
	GetID() string
	SyncContacts() (bool, error)
	SyncChats() (bool, error)
}

type TG interface {
	SendQR(mgChatID int64, code string) (msg *TGMessage, err error)
	SendMessage(chatID int64, text string) (msg *TGMessage, err error)
	SendImage(chatID int64, reader io.Reader, caption string) (msg *TGMessage, err error)
	SendAudio(chatID int64, reader io.Reader) (msg *TGMessage, err error)
	SendVideo(chatID int64, reader io.Reader) (msg *TGMessage, err error)
	SendDocument(chatID int64, reader io.Reader, fileName string) (msg *TGMessage, err error)
	SendLocation(chatID int64, lat, lon float64) (msg *TGMessage, err error)
	DeleteMessage(chatID int64, messageID int) (err error)
	UpdateStatMessage()
	SendLog(text string)
	GetMainGroups() []int64
}

type Store interface {
	GetMainGroupByName(name string) (apiItem *MainGroup, err error)
	GetMainGroupByTGID(id int64) (apiItem *MainGroup, err error)
	SaveMainGroup(mg *MainGroup) (err error)
	SaveMessage(message *Message) error
	GetMessageByWA(messageID string) (*Message, error)
	GetMessagesNotChattedByClient(client string) ([]*Message, error)
	ExistMessageByWA(messageID string) bool
	ExistMessageByTG(messageID int, chatID int64) bool
	GetChatByClient(client string, id string) (*Chat, error)
	GetChatsByChatID(chatID int64) ([]*Chat, error)
	SaveChat(chat *Chat) error
	GetStatOnPeriod(mgChatID int64, userName string, start, end time.Time) (apiItems []*Stat, err error)
	DeleteChat(chat *Chat) (bool, error)
	GetNotChatted(mgID int64, botName string) (apiItems []*StatDay, err error)
	SaveAlias(alias *Alias) (err error)
	GetAliasesByName(name string) (apiItems []*Alias, err error)
	SaveContact(contact *Contact) (err error)
	GetContactsByPhone(phone string) (apiItems []*Contact, err error)
	GetContactsByWAClient(waClient string) (apiItems []*Contact, err error)
}

type Cache interface {
	GetMembers() ([]int, error)
}
