package wa

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"time"

	"github.com/cristalinojr/go-whatsapp"
)

type Service struct {
	instances map[int64]*Instance
	api.WA
}

type StatusItem struct {
	At   time.Time
	Desc string
}

type InstanceStatus struct {
	ChatsLoad    StatusItem
	ContactsLoad StatusItem
}

type Instance struct {
	ctx  context.Context
	id   int64
	conn *whatsapp.Conn
	api.WAInstance
	clients   []string
	pointTime uint64
	status    InstanceStatus
}

func New(ctx context.Context) (service *Service, err error) {

	service = &Service{instances: map[int64]*Instance{}}
	pointTimeStr := os.Getenv("WA_POINT_TIME")
	pointTime, err := strconv.ParseUint(pointTimeStr, 10, 64)
	if err != nil {
		pointTime = 0
	}

	items := strings.Split(os.Getenv("TG_MAIN_GROUPS"), ",")

	for _, v := range items {

		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return service, fmt.Errorf("error parse ID: %w", err)
		}

		instance := &Instance{ctx: ctx, clients: []string{}, id: id, pointTime: pointTime}

		instance.conn, err = whatsapp.NewConn(30 * time.Second)
		if err != nil {
			return service, fmt.Errorf("error creating connection: %w", err)
		}

		instance.conn.SetClientVersion(2, 2142, 12)

		instance.conn.AddHandler(instance)
		if err = instance.login(true); err != nil {
			return service, fmt.Errorf("error login: %w", err)
		}

		instance.WAInstance = instance
		service.instances[id] = instance
	}

	return
}

func (s *Service) UpdateCTX(ctx context.Context) {
	for k := range s.instances {
		s.instances[k].ctx = ctx
	}
}

func (s *Service) ShutDown() error {
	for _, v := range s.instances {
		session, err := v.conn.Disconnect()
		if err != nil {
			return fmt.Errorf("error disconnecting: %w", err)
		}
		if err = v.writeSession(session); err != nil {
			return fmt.Errorf("error saving session: %w", err)
		}
	}
	return nil
}

func (s *Service) GetInstance(id int64) (api.WAInstance, bool) {
	item, ok := s.instances[id]
	return item, ok
}

func (s *Instance) login(onlyRestore bool) error {

	var ok bool
	var err error

	if ok, err = s.restore(); err != nil && !onlyRestore {
		return fmt.Errorf("error restore: %w", err)
	} else if err != nil {
		log.Println("error restore: ", err)
	}

	if !ok {
		if ok, err = s.restoreSession(); err != nil && !onlyRestore {
			return fmt.Errorf("error restore session: %w", err)
		}
	} else if err != nil {
		log.Println("error restore session: ", err)
	}

	if !ok && !onlyRestore {
		if ok, err = s.loginSession(); err != nil {
			return fmt.Errorf("error login session: %w", err)
		}
	}

	if !ok && !onlyRestore {
		return fmt.Errorf("error bad status login ")
	}

	err = s.conn.AdminTest()
	if err != nil {
		if !onlyRestore {
			return fmt.Errorf("error ping: %w", err)
		}
	}

	log.Println("WAInstance Status: ", ok)

	versionServer, err := whatsapp.CheckCurrentServerVersion()
	if err != nil {
		return fmt.Errorf("error set version: %v", err)
	}

	s.conn.SetClientVersion(versionServer[0], versionServer[1], versionServer[2])
	versionClient := s.conn.GetClientVersion()

	log.Printf("whatsapp version %v.%v.%v", versionClient[0], versionClient[1], versionClient[2])

	return nil
}

func (s *Instance) restore() (ok bool, err error) {
	err = s.conn.Restore()
	if err != nil && (errors.Is(err, whatsapp.ErrAlreadyConnected) || errors.Is(err, whatsapp.ErrAlreadyLoggedIn)) {
		return true, nil
	} else if err != nil && (errors.Is(err, whatsapp.ErrInvalidSession)) {
		return false, nil
	}
	return true, nil
}

func (s *Instance) restoreSession() (ok bool, err error) {
	session, err := s.readSession()
	if err == nil && session.ClientId != "" {
		session, err = s.conn.RestoreWithSession(session)
		if err != nil && (errors.Is(err, whatsapp.ErrAlreadyConnected) || errors.Is(err, whatsapp.ErrAlreadyLoggedIn)) {
			return true, nil
		} else if err != nil && (errors.Is(err, whatsapp.ErrInvalidSession)) {
			return false, nil
		}
		if err != nil {
			if errSave := s.writeSession(session); errSave != nil {
				return false, fmt.Errorf("error saving session: %v", errSave)
			}
			return false, fmt.Errorf("restoring session failed: %w. Please try again. ", err)
		}
		return true, nil
	}
	return false, nil
}

func (s *Instance) loginSession() (ok bool, err error) {
	qr := make(chan string)
	tg, _ := appCtx.FromTG(s.ctx)
	var tgMsg *api.TGMessage
	go func() {
		if tg == nil {
			tg, _ = appCtx.FromTG(s.ctx)
		}
		tgMsg, err = tg.SendQR(s.id, <-qr)
		if err != nil {
			log.Printf("error send QR: %v\n", err)
		}
	}()
	session, err := s.conn.Login(qr)
	if tgMsg != nil {
		err = tg.DeleteMessage(tgMsg.ChatID, tgMsg.MessageID)
		if err != nil {
			return false, fmt.Errorf("error delete tg message: %w", err)
		}
	}
	if err != nil {
		return false, fmt.Errorf("error during login: %w", err)
	}
	if errSave := s.writeSession(session); errSave != nil {
		_ = s.conn.Logout()
		_, _ = s.conn.Disconnect()
		return false, fmt.Errorf("error saving session: %v", errSave)
	}
	return true, nil
}

func (s *Instance) readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(fmt.Sprintf("%d_wa_instance_session.gob", s.id))
	if err != nil {
		return session, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("error close file: ", err)
		}
	}()

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&session); err != nil {
		return session, err
	}

	return session, nil
}

func (s *Instance) writeSession(session whatsapp.Session) error {
	file, err := os.Create(fmt.Sprintf("%d_wa_instance_session.gob", s.id))
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("error close file: ", err)
		}
	}()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(session); err != nil {
		return err
	}

	return nil
}
