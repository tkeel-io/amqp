package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tkeel-io/amqp/internal/mq"
	"net/http"
	"strings"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/pkg/errors"
	tkeelAMQP "github.com/tkeel-io/amqp"
	"github.com/tkeel-io/kit/auth"
	"github.com/tkeel-io/kit/log"
)

func NewBroker(addr string) *tkeelAMQP.Broker {
	return tkeelAMQP.NewBroker(addr,
		userAuthHandler,
		senderHandler,
		electron.SASLAllowedMechs("ANONYMOUS"),
		electron.SASLAllowInsecure(true))
}

func userAuthHandler(conn electron.Connection, s electron.Sender) (interface{}, error) {
	token := conn.VirtualHost()
	if token == "" || strings.HasPrefix("Bearer", token) {
		return nil, errors.New("invalid token")
	}
	user, err := auth.Authenticate(token, auth.AuthTokenURLInvoke)
	if err != nil {
		log.Error("auth failed", err)
		return nil, err
	}
	if !validateTopic(user, s.Source()) {
		return nil, errors.New("topic and user mismatch")
	}
	return user, nil
}

// senderHandler process with message witch u want to send
func senderHandler(ctx context.Context, s electron.Sender) <-chan amqp.Message {
	ch := make(chan amqp.Message, 1)
	go func(s electron.Sender, ch chan<- amqp.Message) {
		source := mq.FindSourceChan(s.Source())
		go mq.Connect(context.WithValue(ctx, "id", s.Source()), s.Source())
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case data, ok := <-source:
				if !ok {
					continue
				}
				fmt.Printf("get data:%s\n", data)
				ch <- amqp.NewMessageWith(data)
			}
		}
	}(s, ch)
	return ch
}

func validateTopic(user *auth.User, topic string) bool {
	data := map[string]string{
		"topic": topic,
	}
	content, err := json.Marshal(data)
	if err != nil {
		log.Error("json marshal failed", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, _subscribeValidateURL, bytes.NewReader(content))
	req.Header.Add("Authorization", user.Token)
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Error("POST topic validate error ", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("POST topic validate error ", resp.StatusCode)
		return false
	}
	return true
}
