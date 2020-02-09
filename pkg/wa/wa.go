package wa

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"tgwabr/api"
	appCtx "tgwabr/context"
	"time"

	"github.com/Rhymen/go-whatsapp"
)

type Service struct {
	ctx     context.Context
	conn    *whatsapp.Conn
	clients []string
	api.WA
	pointTime uint64
}

func New(ctx context.Context) (service *Service, err error) {

	service = &Service{ctx: ctx, clients: []string{}}
	pointTimeStr := os.Getenv("WA_POINT_TIME")
	pointTime, err := strconv.ParseUint(pointTimeStr, 10, 64)
	if err == nil {
		service.pointTime = pointTime
	}
	service.conn, err = whatsapp.NewConn(5 * time.Second)
	if err != nil {
		return service, fmt.Errorf("error creating connection: %w", err)
	}

	service.conn.AddHandler(service)

	if err = service.login(true); err != nil {
		return service, fmt.Errorf("error login: %w", err)
	}

	service.WA = service
	return
}

func (s *Service) UpdateCTX(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) ShutDown() error {
	session, err := s.conn.Disconnect()
	if err != nil {
		return fmt.Errorf("error disconnecting: %w", err)
	}
	if err = s.writeSession(session); err != nil {
		return fmt.Errorf("error saving session: %w", err)
	}
	return nil
}

func (s *Service) login(onlyRestore bool) error {

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

	ok, err = s.conn.AdminTest()
	if err != nil {
		if !onlyRestore {
			return fmt.Errorf("error ping: %w", err)
		}
	}

	log.Println("WA Status: ", ok)

	versionServer, err := whatsapp.CheckCurrentServerVersion()
	if err != nil {
		return fmt.Errorf("error set version: %v", err)
	}

	s.conn.SetClientVersion(versionServer[0], versionServer[1], versionServer[2])
	versionClient := s.conn.GetClientVersion()

	log.Printf("whatsapp version %v.%v.%v", versionClient[0], versionClient[1], versionClient[2])

	return nil
}

func (s *Service) restore() (ok bool, err error) {
	err = s.conn.Restore()
	if err != nil && (errors.Is(err, whatsapp.ErrAlreadyConnected) || errors.Is(err, whatsapp.ErrAlreadyLoggedIn)) {
		return true, nil
	} else if err != nil && (errors.Is(err, whatsapp.ErrInvalidSession)) {
		return false, nil
	}
	return true, nil
}

func (s *Service) restoreSession() (ok bool, err error) {
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

func (s *Service) loginSession() (ok bool, err error) {
	qr := make(chan string)
	tg, _ := appCtx.FromTG(s.ctx)
	var tgMsg *api.TGMessage
	go func() {
		if tg == nil {
			tg, _ = appCtx.FromTG(s.ctx)
		}
		tgMsg, err = tg.SendQR(<-qr)
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

func (s *Service) readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	name := os.Getenv("NAME_INSTANCE")
	file, err := os.Open(name + "_WASession.gob")
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

func (s *Service) writeSession(session whatsapp.Session) error {
	name := os.Getenv("NAME_INSTANCE")
	file, err := os.Create(name + "_WASession.gob")
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
