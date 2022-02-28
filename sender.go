package amqp

import (
	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/tkeel-io/kit/log"
	"strings"
)

type SenderManager struct {
	Container *electron.Container
	Conn      map[string]electron.Connection
	Sender    map[string]electron.Sender
	Outcome   chan electron.Outcome
}

func (s *SenderManager) Run(urls map[string]string) error {
	for id, urlStr := range urls {
		url, err := amqp.ParseURL(urlStr)
		if err != nil {
			log.Error("error parsing url", err)
			return err
		}
		conn, err := electron.Dial("tcp", url.Host)
		if err != nil {
			return err
		}
		s.Conn[id] = conn
		addr := strings.TrimPrefix(url.Path, "/")
		sender, err := conn.Sender(electron.Target(addr))
		if err != nil {
			log.Error("error creating sender", err)
			return err
		}
		s.Sender[id] = sender
	}
	return nil
}

func (s *SenderManager) Send(id string, body []byte) {
	sender, ok := s.Sender[id]
	if !ok {
		return
	}
	m := amqp.NewMessage()
	m.Marshal(body)
	sender.SendAsync(m, s.Outcome, body)
}

func (s *SenderManager) Close(ids ...string) error {
	for i := range ids {
		if conn, ok := s.Conn[ids[i]]; ok {
			conn.Close(nil)
			delete(s.Conn, ids[i])
		}
		if sender, ok := s.Sender[ids[i]]; ok {
			sender.Close(nil)
			delete(s.Conn, ids[i])
		}
	}
	return nil
}

func (s *SenderManager) Shutdown() {
	for id, conn := range s.Conn {
		conn.Close(nil)
		delete(s.Conn, id)
	}
	for id, sender := range s.Sender {
		sender.Close(nil)
		delete(s.Sender, id)
	}
}
