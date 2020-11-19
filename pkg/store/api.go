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

func (s *Store) GetStatOnPeriod(mgChatID int64, userName string, start, end time.Time) (res []*api.Stat, err error) {
	res = []*api.Stat{}
	q := s.db.Table("messages").Select(`
       date(created_at)                                   "date",
       tg_user_name,
       wa_client,
       wa_name,
       min(case when session != '' then created_at end)   "session",
       min(case when answered > 0 then answered end) / 60 "answered",
       count(case when direction = 'wa2tg' then id end)   "count_in",
       count(case when direction = 'tg2wa' then id end)   "count_out"`).
		Where("mg_id = ? and created_at between ? and DATE_ADD(?, INTERVAL 1 DAY)", mgChatID, start, end)
	if userName != "" {
		q = q.Where("tg_user_name = ?", userName)
	}
	q = q.Group("date(created_at), tg_user_name, session, wa_name, wa_client").
		Order("date").
		Order("tg_user_name").
		Order("min(case when session != '' then created_at end)").
		Order("wa_name").
		Order("wa_client")

	err = q.Scan(&res).Error

	return res, err
}

func (s *Store) DeleteChat(chat *api.Chat) (bool, error) {
	item := &Chat{}
	ok, err := s.FindOne(s.db.Model(&Chat{}).Where(&Chat{
		TGChatID: chat.TGChatID,
	}), item)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	err = s.db.Unscoped().Delete(&item).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) GetNotChatted(mgID int64, botName string) (res []*api.StatDay, err error) {
	res = []*api.StatDay{}
	err = s.db.Raw(`
	select distinct messages.created_at "date",
       t0.wa_client,
       (select tg_user_name
        from messages
        where chatted = 'yes'
          and mg_id = t0.mg_id
          and wa_client = t0.wa_client
          and tg_user_name != ?
        order by created_at DESC
        limit 1)           tg_user_name,
       t0.cnt              "count",
       t0.cnt_unr              "count_unread"
	from (select max(created_at)      at,
				 wa_client,
				 count(wa_message_id) cnt,
				 count(case when message_status < 4 then wa_message_id end) cnt_unr,	
				 mg_id
		  from messages
		  where chatted = 'no'
			and mg_id = ?
		  group by wa_client, tg_user_name, tg_user_name, mg_id) t0
			 inner join messages
						on t0.wa_client = messages.wa_client
							and t0.at = messages.created_at
							and t0.mg_id = messages.mg_id
	order by
		messages.created_at asc;
	`, botName, mgID).Scan(&res).Error
	return
}

func (s *Store) SaveAlias(alias *api.Alias) (err error) {

	item := &Alias{}
	_, err = s.FindOne(s.db.Model(&Alias{}).Where(&Alias{WAClient: alias.WAClient, MGID: alias.MGID}), item)
	if err != nil {
		return err
	}
	id := item.ID
	item = APIAlias(*alias).ToAlias()
	item.ID = id
	err = s.db.Save(item).Error
	if err != nil {
		return err
	}
	return
}

func (s *Store) GetAliasesByName(name string) (apiItems []*api.Alias, err error) {

	items := Aliases{}
	err = s.db.Model(&Alias{}).Find(&items, &Alias{Name: name}).Error
	if err != nil {
		return
	}
	return items.ToAPIAliases(), nil
}

func (s *Store) GetAliasesByWAClient(waClient string) (apiItems []*api.Alias, err error) {

	items := Aliases{}
	err = s.db.Model(&Alias{}).Find(&items, &Alias{WAClient: waClient}).Error
	if err != nil {
		return
	}
	return items.ToAPIAliases(), nil
}

func (s *Store) SaveContact(contact *api.Contact) (err error) {

	item := &Contact{}
	_, err = s.FindOne(s.db.Model(&Contact{}).Where(&Contact{Phone: contact.Phone}), item)
	if err != nil {
		return err
	}
	id := item.ID
	item = APIContact(*contact).ToContact()
	item.ID = id
	err = s.db.Save(item).Error
	if err != nil {
		return err
	}
	return
}

func (s *Store) GetContactsByPhone(phone string) (apiItems []*api.Contact, err error) {

	items := Contacts{}
	err = s.db.Model(&Contact{}).Find(&items, &Contact{Phone: phone}).Error
	if err != nil {
		return
	}
	return items.ToAPIContacts(), nil
}
func (s *Store) GetContactsByWAClient(waClient string) (apiItems []*api.Contact, err error) {

	items := Contacts{}
	err = s.db.Model(&Contact{}).Find(&items, &Contact{WAClient: waClient}).Error
	if err != nil {
		return
	}
	return items.ToAPIContacts(), nil
}
