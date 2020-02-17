package store

import (
	"log"
	"tgwabr/api"
	"time"
)

func (s *Store) SaveMessage(message *api.Message) (err error) {
	item := &Message{}
	condition := &Message{WAMessageID: message.WAMessageID}
	if message.WAMessageID == "" {
		condition = &Message{TGMessageID: message.TGMessageID, TGChatID: message.TGChatID}
	}
	_, err = s.FindOne(s.db.Model(&Message{}).Where(condition), item)
	if err != nil {
		return err
	}
	id := item.ID
	item = APIMessage(*message).ToMessage()
	item.ID = id
	err = s.db.Save(item).Error
	if err != nil {
		return err
	}
	return
}

func (s *Store) SaveChat(chat *api.Chat) (err error) {

	item := &Chat{}
	_, err = s.FindOne(s.db.Model(&Chat{}).Where(&Chat{WAClient: chat.WAClient}), item)
	if err != nil {
		return err
	}
	id := item.ID
	item = APIChat(*chat).ToChat()
	item.ID = id
	err = s.db.Save(item).Error
	if err != nil {
		return err
	}
	return
}

func (s *Store) SaveMainGroup(mg *api.MainGroup) (err error) {

	item := &MainGroup{}
	_, err = s.FindOne(s.db.Model(&MainGroup{}).Where(&MainGroup{TGChatID: mg.TGChatID}), item)
	if err != nil {
		return err
	}
	id := item.ID
	item = APIMainGroup(*mg).ToMainGroup()
	item.ID = id
	err = s.db.Save(item).Error
	if err != nil {
		return err
	}
	return
}

func (s *Store) GetMainGroupByName(name string) (apiItem *api.MainGroup, err error) {

	item := &MainGroup{}
	ok, err := s.FindOne(s.db.Model(&MainGroup{}).Where(&MainGroup{Name: name}), item)
	if err != nil {
		return
	}
	if !ok {
		return nil, nil
	}
	return item.ToAPIMainGroup(), nil
}

func (s *Store) GetMainGroupByTGID(id int64) (apiItem *api.MainGroup, err error) {

	item := &MainGroup{}
	ok, err := s.FindOne(s.db.Model(&MainGroup{}).Where(&MainGroup{TGChatID: id}), item)
	if err != nil {
		return
	}
	if !ok {
		return nil, nil
	}
	return item.ToAPIMainGroup(), nil
}

func (s *Store) GetMessageByWA(messageID string) (apiItem *api.Message, err error) {
	item := &Message{}
	ok, err := s.FindOne(s.db.Model(&Message{}).Where(&Message{WAMessageID: messageID}), item)
	if err != nil {
		return
	}
	if !ok {
		return nil, nil
	}
	return item.ToAPIMessage(), nil
}

func (s *Store) ExistMessageByWA(messageID string) bool {
	ok, err := s.Check(s.db.Model(&Message{}).Where(&Message{WAMessageID: messageID}))
	if err != nil {
		log.Println("Error ExistMessageByWA: ", err)
	}
	return ok
}

func (s *Store) ExistMessageByTG(messageID int, chatID int64) bool {
	ok, err := s.Check(s.db.Model(&Message{}).Where(&Message{TGMessageID: messageID, TGChatID: chatID}))
	if err != nil {
		log.Println("Error ExistMessageByTG: ", err)
	}
	return ok
}

func (s *Store) GetChatByClient(client string, id string) (chat *api.Chat, err error) {
	item := &Chat{}
	ok, err := s.FindOne(s.db.Model(&Chat{}).Where(&Chat{WAClient: client, MGID: id}), item)
	if err != nil {
		return
	}
	if !ok {
		return nil, nil
	}

	return item.ToAPIChat(), nil
}

func (s *Store) GetChatsByChatID(chatID int64) (chats []*api.Chat, err error) {

	items := Chats{}
	err = s.db.Model(&Chat{}).Find(&items, &Chat{TGChatID: chatID}).Error
	if err != nil {
		return
	}
	return items.ToAPIChats(), nil
}

func (s *Store) GetMessagesNotChattedByClient(client string) (msg []*api.Message, err error) {

	items := Messages{}
	err = s.db.Model(&Message{}).Find(&items, &Message{Chatted: api.ChattedNo, WAClient: client}).Error
	if err != nil {
		return
	}
	return items.ToAPIMessages(), nil
}

func (s *Store) GetStatOnPeriod(mgChatID int64, userName string, start, end time.Time) ([]*api.Stat, error) {
	return nil, nil
}

func (s *Store) DeleteChat(chat *api.Chat) (bool, error) {
	item, err := s.GetChatByClient(chat.WAClient, chat.MGID)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	err = s.db.Delete(item).Error
	if err != nil {
		return false, nil
	}
	return true, nil
}
