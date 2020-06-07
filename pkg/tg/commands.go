package tg

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"tgwabr/api"
	"tgwabr/context"
	"time"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (s *Service) CommandRestart(update tgBotApi.Update) {
	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			msg.Text = msg.Text + "\n #restart"
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	msg.Text = "Service restarted..."
	os.Exit(0)
}

func (s *Service) CommandAutoReplay(update tgBotApi.Update) {
	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	///autoreplay
	//*;
	//17:30-8:00|1-7|*|*;
	//Hello!

}

func (s *Service) CommandSync(update tgBotApi.Update) {
	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	if ok, err := wac.SyncContacts(); err != nil {
		msg.Text = fmt.Sprintf("Sync contact ERROR: %s", err.Error())
	} else if ok {
		msg.Text = fmt.Sprintf("Sync contact OK")
	} else {
		msg.Text = fmt.Sprintf("Sync contact BAD")
	}

	if ok, err := wac.SyncChats(); err != nil {
		msg.Text = fmt.Sprintf("%s, Sync chat ERROR: %s", msg.Text, err.Error())
	} else if ok {
		msg.Text = fmt.Sprintf("%s, Sync chat OK", msg.Text)
	} else {
		msg.Text = fmt.Sprintf("%s, Sync chat BAD", msg.Text)
	}
}

func (s *Service) CommandRePined(update tgBotApi.Update) {
	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	grp, err := db.GetMainGroupByTGID(chatID)
	if err != nil {
		log.Println("Error get MainGroup updateStatMessage: ", err)
		return
	}

	if grp == nil {
		log.Println("Error get MainGroup not found: ", chatID)
		return
	}

	grp.MessagePin = -1
	_ = db.SaveMainGroup(grp)

	s.UpdateStatMessage()
}

func (s *Service) CommandCheckClient(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if s.IsMainGroup(chatID) {
		msg.Text = "Main group not check client"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	args := update.Message.CommandArguments()
	args = strings.ToLower(strings.TrimSpace(args))
	client, _ := s.prepareArgs(args)
	client = s.prepareClient(client)
	txt := fmt.Sprintf("Check client: %s", client)
	isFound := false
	for _, v := range s.mainGroups {

		wac, ok := waSvc.GetInstance(v)
		if !ok {
			msg.Text = "Instance WhatsApp not ready"
			return
		}

		if !wac.ClientExist(client) {
			continue
		}
		isFound = true

		mg, _ := db.GetMainGroupByTGID(v)
		mgName := "-"
		if mg != nil {
			mgName = mg.Name
		}
		txt = fmt.Sprintf("%s\n - %s, JID: %s, name: %s, mg: %s", txt, client, wac.PrepareClientJID(client), wac.GetClientName(client), mgName)
	}
	if !isFound {
		txt = txt + " not found"
	}

	msg.Text = txt
}

func (s *Service) CommandStat(update tgBotApi.Update) {
	chatID := update.Message.Chat.ID
	var err error
	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
		if err != nil {
			s.SendLog(err.Error())
		}
	}()

	if s.IsMainGroup(chatID) {
		msg.Text = "Command not work in Main group"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	args := update.Message.CommandArguments()
	args = strings.ToLower(strings.TrimSpace(args))
	argItems := strings.Split(args, " ")
	dateStart := time.Now()
	dateEnd := dateStart
	if len(argItems) > 0 && args != "" {
		dateStart, err = time.Parse("2006-01-02", argItems[0])
		if err != nil {
			msg.Text = fmt.Sprintf("Fail parse start date. Please input date on format YYYY-MM-DD, please send admin this error: %s", err)
			return
		}
		dateEnd = dateStart
	}
	if len(argItems) > 1 {
		dateEnd, err = time.Parse("2006-01-02", argItems[1])
		if err != nil {
			msg.Text = fmt.Sprintf("Fail parse end date. Please input date on format YYYY-MM-DD, please send admin this error: %s", err)
			return
		}
	}

	items := []*api.Stat{}
	for _, v := range s.mainGroups {
		var member tgBotApi.ChatMember
		member, err = s.bot.GetChatMember(tgBotApi.ChatConfigWithUser{
			ChatID: v,
			UserID: update.Message.From.ID,
		})

		if err != nil {
			msg.Text = fmt.Sprintf("Fail get member of main group, please send admin this error: %s", err)
			return
		}

		userName := ""
		if !(member.IsCreator() || member.IsAdministrator()) {
			userName = update.Message.From.UserName
		}
		var res []*api.Stat
		res, err = db.GetStatOnPeriod(v, userName, dateStart, dateEnd)
		if err != nil {
			msg.Text = fmt.Sprintf("Fail get Stat, please send admin this error: %s", err)
			return
		}
		items = append(items, res...)
	}
	txt := ""
	for _, v := range items {
		if v == nil {
			continue
		}
		txt = fmt.Sprintf("%s\n%s;%s;%s;%d", txt, v.Date.Format("2006-01-02"), v.TGUserName, v.WAName, v.Count)
	}
	if txt == "" {
		txt = "Stat not found from period"
	}
	msg.Text = txt
}

func (s *Service) CommandSet(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	member, err := s.bot.GetChatMember(tgBotApi.ChatConfigWithUser{
		ChatID: chatID,
		UserID: update.Message.From.ID,
	})

	if err != nil {
		msg.Text = fmt.Sprintf("Fail get member of main group, please send admin this error: %s", err)
		return
	}

	if !(member.IsCreator() || member.IsAdministrator()) {
		msg.Text = fmt.Sprintf("Forbbiden, only Admin or Owner")
		return
	}

	params := update.Message.CommandArguments()
	params = strings.ToLower(strings.TrimSpace(params))
	if params == "" {
		msg.Text = fmt.Sprintf("Name MainGroup is empty")
		return
	}

	err = db.SaveMainGroup(&api.MainGroup{
		TGChatID: chatID,
		Name:     params,
	})
	if err != nil {
		msg.Text = fmt.Sprintf("Fail set '%s', please send admin this error: %s", params, err)
		log.Println("Error save mainGroup store: ", err)
		return
	}

	msg.Text = "MainGroup Set: OK"
}

func (s *Service) CommandSetLogger(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if s.IsMainGroup(chatID) {
		msg.Text = "Command not work 'Main group'"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	member, err := s.bot.GetChatMember(tgBotApi.ChatConfigWithUser{
		ChatID: chatID,
		UserID: update.Message.From.ID,
	})

	if err != nil {
		msg.Text = fmt.Sprintf("Fail get member of main group, please send admin this error: %s", err)
		return
	}

	if !(member.IsAdministrator()) {
		msg.Text = fmt.Sprintf("Forbbiden, only Admin")
		return
	}

	mgName := update.Message.CommandArguments()
	mgName = strings.ToLower(strings.TrimSpace(mgName))

	mg, err := db.GetMainGroupByName(mgName)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get MainGroup '%s', please send admin this error: %s", mgName, err)
		log.Println("Error get mainGroup store: ", err)
		return
	}
	if mg == nil {
		msg.Text = fmt.Sprintf("Fail, MainGroup '%s' not found", mgName)
		return
	}
	if !s.IsMemberMainGroup(update.Message.From.ID, mg.TGChatID) {
		msg.Text = fmt.Sprintf("Access denied! You are not MainGroup '%s' member", mgName)
		return
	}

	mg.LoggerChatID = chatID

	err = db.SaveMainGroup(mg)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail set logger for '%s', please send admin this error: %s", mgName, err)
		log.Println("Error save mainGroup store: ", err)
		return
	}

	msg.Text = "Set Logger: OK"
}

func (s *Service) CommandStatus(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	device := "Disconnect"
	if wac.GetStatusDevice() {
		device = "Connect"
	}

	loadContactStr := "No"
	loadContact, countContacts, descContacts := wac.GetStatusContacts()
	if loadContact {
		loadContactStr = "Yes"
	}

	login := "Offline"
	if wac.GetStatusLogin() {
		login = "Online"
	}

	chatStat := ""
	chats, countChats, descChats := wac.GetUnreadChat()
	i := 0
	for k, v := range chats {
		if i > 4 {
			break
		}
		chatStat = chatStat + fmt.Sprintf(" - %s (%s): %s\n", wac.GetClientName(k), wac.GetShortClient(k), v)
		i++
	}
	if i > 0 && len(chats) > i {
		chatStat = chatStat + fmt.Sprintf(" ... %d more chats\n", len(chats))
	}
	if chatStat == "" {
		chatStat = "No unread chats :)"
	}

	msg.Text = fmt.Sprintf(`
Device: %s
Login: %s
Contacts: %s, load: %s, count: %d
Chats: %s, count: %d, items:
 %s
`, device, login, descContacts, loadContactStr, countContacts, descChats, countChats, chatStat)
}

func (s *Service) CommandHistory(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	items, err := db.GetChatsByChatID(chatID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get History chat, please send admin this error: %s", err)
		log.Println("Error save chat store: ", err)
		return
	}

	if len(items) == 0 {
		msg.Text = fmt.Sprintf("Chat not joined!")
		return
	}
	mgChatID, err := strconv.ParseInt(items[0].MGID, 10, 64)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get History, please send admin this error: %s", err)
		log.Println("Error parse MGID: ", err)
		return
	}
	wac, ok := waSvc.GetInstance(mgChatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	client := items[0].WAClient
	name := wac.GetClientName(client)

	params := update.Message.CommandArguments()
	params = strings.ToLower(strings.TrimSpace(params))
	size := 0
	if size, err = strconv.Atoi(params); err != nil {
		size = 10
	}

	err = wac.GetHistory(client, size)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get History chat for '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error get History: ", err)
	}

	msg.Text = ""
}

func (s *Service) CommandLogin(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	ok, err := wac.DoLogin()
	if err != nil {
		msg.Text = fmt.Sprintf("Error login: %s", err)
	} else if ok {
		msg.Text = "Login OK"
	} else {
		msg.Text = "Login FAIL"
	}
}

func (s *Service) CommandAlias(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	args := update.Message.CommandArguments()
	args = strings.ToLower(strings.TrimSpace(args))

	client, aliasName := s.prepareArgs(args)
	client = s.prepareClient(client)
	if client == "" {
		msg.Text = "Client not set"
		return
	}
	if aliasName == "" {
		msg.Text = "Alias not set"
		return
	}

	if !wac.ClientExist(client) {
		msg.Text = fmt.Sprintf("Client '%s' not found", client)
		return
	}

	alias := &api.Alias{
		MGID:     fmt.Sprintf("%d", chatID),
		WAClient: client,
		Name:     aliasName,
	}

	err := db.SaveAlias(alias)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail save alias '%s' - '%s', please send admin this error: %s", client, aliasName, err)
		log.Println("Error save chat store: ", err)
		return
	}

	msg.Text = fmt.Sprintf("Client '%s' save as '%s'", client, aliasName)
}

func (s *Service) CommandContact(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	args := update.Message.CommandArguments()
	args = strings.ToLower(strings.TrimSpace(args))

	client, contactName := s.prepareArgs(args)
	client = s.prepareClient(client)
	if client == "" {
		msg.Text = "Client not set"
		return
	}
	if contactName == "" {
		msg.Text = "Contact name not set"
		return
	}

	if !wac.ClientExist(client) {
		msg.Text = fmt.Sprintf("Client '%s' not found", client)
		return
	}

	contact := &api.Contact{
		Phone: client,
		Name:  contactName,
	}

	err := db.SaveContact(contact)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail save alias '%s' - '%s', please send admin this error: %s", client, contactName, err)
		log.Println("Error save chat store: ", err)
		return
	}

	msg.Text = fmt.Sprintf("Client '%s' save as '%s'", client, contactName)
}

func (s *Service) CommandLogout(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if !s.IsMainGroup(chatID) {
		msg.Text = "Command work only 'Main group'"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	wac, ok := waSvc.GetInstance(chatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	ok, err := wac.DoLogout()
	if err != nil {
		msg.Text = fmt.Sprintf("Error login: %s", err)
	} else if ok {
		msg.Text = "Logout OK"
	} else {
		msg.Text = "Logout FAIL"
	}
}

func (s *Service) CommandJoin(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID
	userName := update.Message.From.UserName
	userID := update.Message.From.ID

	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	if s.IsMainGroup(chatID) {
		msg.Text = "Main group not join client"
		return
	}

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	args := update.Message.CommandArguments()
	args = strings.ToLower(strings.TrimSpace(args))

	client, mgName := s.prepareArgs(args)
	client = s.prepareClient(client)

	if client == "" {
		msg.Text = "Client not set"
		return
	}
	if client == "all" {
		msg.Text = "ALL not work :'("
		return
	}

	aliases, err := db.GetAliasesByName(client)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail get Alias '%s', please send admin this error: %s", client, err)
		log.Println("Error get Alias store: ", err)
		return
	}

	mgChatID := int64(0)
	if len(aliases) == 1 {
		mgChatID, _ = strconv.ParseInt(aliases[0].MGID, 10, 64)
		if !s.IsMemberMainGroup(userID, mgChatID) {
			mgChatID = 0
		}
	}

	if mgName != "" && mgChatID == 0 {
		mg, err := db.GetMainGroupByName(mgName)
		if err != nil {
			msg.Text = fmt.Sprintf("Fail get MainGroup '%s', please send admin this error: %s", mgName, err)
			log.Println("Error get mainGroup store: ", err)
			return
		}
		if mg == nil {
			msg.Text = fmt.Sprintf("Fail, MainGroup '%s' not found", mgName)
			return
		}
		if !s.IsMemberMainGroup(userID, mg.TGChatID) {
			msg.Text = fmt.Sprintf("Access denied! You are not MainGroup '%s' member", mgName)
			return
		}
		mgChatID = mg.TGChatID
	} else if mgChatID == 0 {

		isOne := true
		for _, v := range s.mainGroups {
			isMember := s.IsMemberMainGroup(userID, v)
			if isMember && !isOne {
				msg.Text = fmt.Sprintf("Fail, You are part of severall MainGroups, please specify the one. Example: /join tel[or alias] group")
				return
			}
			if isMember && isOne {
				isOne = false
				mgChatID = v
			}
		}
	}

	wac, ok := waSvc.GetInstance(mgChatID)
	if !ok {
		msg.Text = "Instance WhatsApp not ready"
		return
	}

	if client != "check" && !wac.ClientExist(client) {

		for _, v := range aliases {
			if v.MGID == fmt.Sprintf("%d", mgChatID) {
				client = v.WAClient
			}
		}
	}

	if client != "check" && !wac.ClientExist(client) {
		msg.Text = fmt.Sprintf("Client '%s' not found", client)
		return
	}

	name := wac.GetClientName(client)

	chat := api.Chat{
		MGID:       wac.GetID(),
		WAClient:   wac.PrepareClientJID(client),
		TGChatID:   chatID,
		TGUserName: userName,
	}

	items, err := db.GetChatsByChatID(chat.TGChatID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail join chat '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error save chat store: ", err)
		return
	}

	if len(items) > 0 {
		name := wac.GetClientName(items[0].WAClient)
		msg.Text = fmt.Sprintf("Chat already joined to client '%s(%s)'", name, items[0].WAClient)
		return
	}

	err = db.SaveChat(&chat)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail join chat '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error save chat store: ", err)
		return
	}

	msgJoin := tgBotApi.NewMessage(mgChatID, fmt.Sprintf("Chat %s(%s) join to @%s", name, client, update.Message.From.UserName))
	_, _ = s.BotSend(msgJoin)

	_, _ = s.bot.SetChatTitle(tgBotApi.SetChatTitleConfig{
		ChatID: chat.TGChatID,
		Title:  fmt.Sprintf("Chat with %s(%s)", name, client),
	})

	raw, _ := wac.GetContactPhoto(client)
	if raw != "" {
		resp, err := s.bot.SetChatPhoto(tgBotApi.SetChatPhotoConfig{
			BaseFile: tgBotApi.BaseFile{
				BaseChat: tgBotApi.BaseChat{
					ChatID: chat.TGChatID,
				},
				File: tgBotApi.FileBytes{
					Bytes: getPhotoByte(raw),
				},
			},
		})
		if err != nil {
			log.Println(err)
		}
		log.Println(resp)
	}

	messages, err := db.GetMessagesNotChattedByClient(chat.WAClient)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail join chat '%s(%s)', please send admin this error: %s", name, client, err)
		log.Println("Error get not chatted messages store: ", err)
		return
	}

	for _, v := range messages {

		msgTransfer := tgBotApi.NewMessage(chatID, v.Text)
		err = wac.ReadMessage(v.WAClient, v.WAMessageID)
		if err != nil {
			log.Println("Error transfer message: ", err)
		}
		resp, _ := s.BotSend(msgTransfer)
		err = s.DeleteMessage(v.TGChatID, v.TGMessageID)
		if err != nil {
			log.Println("Error transfer message: ", err)
		}

		if resp.Chat != nil {
			tgMsg := Message(resp).ToAPIMessage()
			v.TGChatID = tgMsg.ChatID
			v.TGMessageID = tgMsg.MessageID
			v.TGTimestamp = tgMsg.Timestamp
			v.TGFwdMessageID = tgMsg.FwdMessageID
		}

		v.TGUserName = userName
		v.Chatted = api.ChattedYes

		err = db.SaveMessage(v)
		if err != nil {
			log.Println("Error save transfer message store: ", err)
		}
	}

	msg.Text = fmt.Sprintf("Join '%s(%s)' OK", name, client)
	s.UpdateStatMessage()
}

func (s *Service) CommandLeave(update tgBotApi.Update) {

	chatID := update.Message.Chat.ID

	var err error
	msg := tgBotApi.NewMessage(chatID, "")
	defer func() {
		if msg.Text != "" {
			_, _ = s.BotSend(msg)
		}
	}()

	waSvc, ok := context.FromWA(s.ctx)
	if !ok {
		msg.Text = "Module WhatsApp not ready"
		return
	}

	db, ok := context.FromDB(s.ctx)
	if !ok {
		msg.Text = "Module Store not ready"
		return
	}

	chats, err := db.GetChatsByChatID(chatID)
	if err != nil {
		msg.Text = fmt.Sprintf("Fail leave 'all' chats, please send admin this error: %s", err)
		log.Println("Error get chats store: ", err)
		return
	}
	txt := "Leave chats: \n"
	for _, v := range chats {
		mgChatID, err := strconv.ParseInt(v.MGID, 10, 64)
		if err != nil {
			msg.Text = fmt.Sprintf("Fail leave 'all' chats, please send admin this error: %s", err)
			log.Println("Error parse MGID: ", err)
			return
		}
		wac, ok := waSvc.GetInstance(mgChatID)
		if !ok {
			msg.Text = "Instance WhatsApp not ready"
			return
		}
		name := wac.GetClientName(v.WAClient)
		_, err = db.DeleteChat(v)
		if err != nil {
			msg.Text = fmt.Sprintf("Fail leave '%s(%s)' chat, please send admin this error: %s", name, v.WAClient, err)
			log.Println("Error delete chat store: ", err)
			return
		}
		txt = txt + fmt.Sprintf(" - '%s(%s)' OK\n", name, v.WAClient)
		msgJoin := tgBotApi.NewMessage(mgChatID, fmt.Sprintf("@%s leave chat %s(%s)", update.Message.From.UserName, name, wac.GetShortClient(v.WAClient)))
		_, _ = s.BotSend(msgJoin)
	}
	msg.Text = txt

	if !s.IsMainGroup(chatID) {
		_, _ = s.bot.SetChatTitle(tgBotApi.SetChatTitleConfig{
			ChatID: update.Message.Chat.ID,
			Title:  fmt.Sprintf("H.W.Bot Free chat"),
		})

		_, _ = s.bot.DeleteChatPhoto(tgBotApi.DeleteChatPhotoConfig{ChatID: update.Message.Chat.ID})
	}
}

func getPhotoByte(path string) []byte {
	resp, err := http.Get(path)
	if err != nil {
		log.Println(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return b
}

func (s *Service) prepareArgs(args string) (arg1, arg2 string) {

	for _, v := range []string{`^([^a-zA-Z]+)$`, `^([^a-zA-Z]+)([A-Za-z].*)$`, `^(\S*)$`, `^(\S*)\s*(\S*)$`} {
		compRegEx := regexp.MustCompile(v)
		match := compRegEx.FindStringSubmatch(args)
		if len(match) > 1 {
			arg1 = match[1]
		}

		if len(match) > 2 {
			arg2 = match[2]
		}

		if len(match) > 1 {
			break
		}
	}

	return
}

func (s *Service) prepareClient(arg string) (client string) {

	client = strings.ReplaceAll(arg, " ", "")

	if strings.Count(client, "-") == 1 {
		parts := strings.Split(client, "-")
		if _, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			if _, err = strconv.ParseInt(parts[1], 10, 64); err == nil {
				return
			}
		}
	}

	client = strings.ReplaceAll(client, "(", "")
	client = strings.ReplaceAll(client, ")", "")
	client = strings.ReplaceAll(client, "-", "")
	client = strings.ReplaceAll(client, "+", "")
	return
}
