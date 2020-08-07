package store

import (
	"context"
	"fmt"
	"os"
	"tgwabr/api"
	"tgwabr/pkg"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type MainGroup struct {
	gorm.Model

	TGChatID     int64  `gorm:"index"`
	Name         string `gorm:"index"`
	MessagePin   int
	LoggerChatID int64
}

type Chat struct {
	gorm.Model

	MGID       string `gorm:"index"`
	WAClient   string `gorm:"index"`
	TGChatID   int64  `gorm:"index"`
	TGUserName string
	Session    string
}

type Message struct {
	gorm.Model

	MGID           string `gorm:"index"`
	WAClient       string `gorm:"index"`
	WAName         string
	WAFromClient   string
	WAFromName     string
	WAMessageID    string `gorm:"index"`
	WATimestamp    uint64
	WAFwdMessageID string `gorm:"index"`
	TGChatID       int64  `gorm:"index"`
	TGUserName     string
	TGMessageID    int `gorm:"index"`
	TGTimestamp    int
	TGFwdMessageID int    `gorm:"index"`
	Direction      string `gorm:"index"`
	Chatted        string `gorm:"index"`
	Answered       uint64
	MessageStatus  int
	Text           string
	Session        string `gorm:"index"`
}

type Alias struct {
	gorm.Model

	MGID     string
	WAClient string `gorm:"index"`
	Name     string `gorm:"index"`
}

type Contact struct {
	gorm.Model

	Phone     string
	Email     string
	WAClient  string `gorm:"index"`
	TGUserID  int    `gorm:"index"`
	Name      string `gorm:"index"`
	ShortName string
}

type APIMessage api.Message

func (a APIMessage) ToMessage() *Message {
	item := &Message{}
	pkg.MustCopyValue(item, &a)
	return item
}

func (a Message) ToAPIMessage() *api.Message {
	item := &api.Message{}
	pkg.MustCopyValue(item, &a)
	return item
}

type Messages []*Message

func (a Messages) ToAPIMessages() []*api.Message {
	list := make([]*api.Message, len(a))
	for i, item := range a {
		list[i] = item.ToAPIMessage()
	}
	return list
}

type APIChat api.Chat

func (a APIChat) ToChat() *Chat {
	item := &Chat{}
	pkg.MustCopyValue(item, &a)
	return item
}

func (a Chat) ToAPIChat() *api.Chat {
	item := &api.Chat{}
	pkg.MustCopyValue(item, &a)
	return item
}

type Chats []*Chat

func (a Chats) ToAPIChats() []*api.Chat {
	list := make([]*api.Chat, len(a))
	for i, item := range a {
		list[i] = item.ToAPIChat()
	}
	return list
}

type APIMainGroup api.MainGroup

func (a APIMainGroup) ToMainGroup() *MainGroup {
	item := &MainGroup{}
	pkg.MustCopyValue(item, &a)
	return item
}

func (a MainGroup) ToAPIMainGroup() *api.MainGroup {
	item := &api.MainGroup{}
	pkg.MustCopyValue(item, &a)
	return item
}

type MainGroups []*MainGroup

func (a MainGroups) ToAPIMainGroups() []*api.MainGroup {
	list := make([]*api.MainGroup, len(a))
	for i, item := range a {
		list[i] = item.ToAPIMainGroup()
	}
	return list
}

type APIAlias api.Alias

func (a APIAlias) ToAlias() *Alias {
	item := &Alias{}
	pkg.MustCopyValue(item, &a)
	return item
}

func (a Alias) ToAPIAlias() *api.Alias {
	item := &api.Alias{}
	pkg.MustCopyValue(item, &a)
	return item
}

type Aliases []*Alias

func (a Aliases) ToAPIAliases() []*api.Alias {
	list := make([]*api.Alias, len(a))
	for i, item := range a {
		list[i] = item.ToAPIAlias()
	}
	return list
}

type APIContact api.Contact

func (a APIContact) ToContact() *Contact {
	item := &Contact{}
	pkg.MustCopyValue(item, &a)
	return item
}

func (a Contact) ToAPIContact() *api.Contact {
	item := &api.Contact{}
	pkg.MustCopyValue(item, &a)
	return item
}

type Contacts []*Contact

func (a Contacts) ToAPIContacts() []*api.Contact {
	list := make([]*api.Contact, len(a))
	for i, item := range a {
		list[i] = item.ToAPIContact()
	}
	return list
}

type Store struct {
	ctx context.Context
	db  *gorm.DB
	api.Store
}

func New(ctx context.Context) (store *Store, err error) {

	store = &Store{ctx: ctx}

	name := os.Getenv("NAME_INSTANCE")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	db := os.Getenv("DB_NAME")
	dialect := os.Getenv("TYPE_DB")

	urn := name + "_main.db"
	if dialect == "mysql" {
		urn = fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, db)
	}
	store.db, err = gorm.Open(dialect, urn)
	if err != nil {
		return store, fmt.Errorf("error open DB: %w", err)
	}

	if os.Getenv("STORE_DEBUG") != "" {
		store.db = store.db.Debug()
	}
	if dialect == "mysql" {
		store.db.DB().SetMaxOpenConns(20)
		store.db.DB().SetMaxIdleConns(2)
		store.db.DB().SetConnMaxLifetime(time.Second * 20)
	}
	// Migrate the schema
	store.db.AutoMigrate(&Chat{})
	store.db.AutoMigrate(&Message{})
	store.db.AutoMigrate(&MainGroup{})
	store.db.AutoMigrate(&Alias{})
	store.db.AutoMigrate(&Contact{})

	return
}

func (s *Store) ShutDown() error {
	return s.db.Close()
}

func (s *Store) FindOne(db *gorm.DB, out interface{}) (bool, error) {
	result := db.First(out)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Check Check if the data exists
func (s *Store) Check(db *gorm.DB) (bool, error) {
	var count int
	result := db.Count(&count)
	if err := result.Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
