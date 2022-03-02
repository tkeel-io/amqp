package service

import (
	"bytes"
	"context"
	"encoding/json"
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
	user, err := auth.Authenticate(token, auth.AuthTokenURLTestRemote)
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
	go func(s electron.Sender) {
		mq.Connect(context.WithValue(ctx, "id", s.Source()), s.Source())
		source := mq.FindSourceChan(s.Source())
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case data := <-source:
				ch <- amqp.NewMessageWith(data)
			}
		}
	}(s)
	return ch
}

func validateTopic(user *auth.User, topic string) bool {
	validateURL := "http://192.168.123.9:30707/apis/core-broker/v1/validate/subscribe"

	data := map[string]string{
		"topic": topic,
	}
	content, err := json.Marshal(data)
	if err != nil {
		log.Error("json marshal failed", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, validateURL, bytes.NewReader(content))
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
