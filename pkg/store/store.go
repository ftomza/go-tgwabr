package store

import (
	"context"
	"fmt"
	"os"
	"tgwabr/api"
	"tgwabr/pkg"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type MainGroup struct {
	gorm.Model

	TGChatID   int64  `gorm:"index"`
	Name       string `gorm:"index"`
	MessagePin int
}

type Chat struct {
	gorm.Model

	MGID     string `gorm:"index"`
	WAClient string `gorm:"index"`
	TGChatID int64  `gorm:"index"`
}

type Message struct {
	gorm.Model

	MGID           string `gorm:"index"`
	WAClient       string `gorm:"index"`
	WAName         string
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
	Text           string
}

type Alias struct {
	gorm.Model

	MGID     string
	WAClient string `gorm:"index"`
	Name     string `gorm:"index"`
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

type Store struct {
	ctx context.Context
	db  *gorm.DB
	api.Store
}

func New(ctx context.Context) (store *Store, err error) {

	store = &Store{ctx: ctx}

	name := os.Getenv("NAME_INSTANCE")

	store.db, err = gorm.Open("sqlite3", name+"_main.db")
	if err != nil {
		return store, fmt.Errorf("error open DB: %w", err)
	}

	if os.Getenv("STORE_DEBUG") != "" {
		store.db = store.db.Debug()
	}
	// Migrate the schema
	store.db.AutoMigrate(&Chat{})
	store.db.AutoMigrate(&Message{})
	store.db.AutoMigrate(&MainGroup{})
	store.db.AutoMigrate(&Alias{})

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
