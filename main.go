package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/auth"
	"github.com/tkeel-io/kit/log"
)

var address string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "amqp-broker",
		Short: "amqp broker starter",
		Long:  "amqp broker starter",
		Run:   run,
	}
	parseFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func parseFlags(root *cobra.Command) {
	root.PersistentFlags().StringVar(&address, "address", ":3172", "the broker address and port you want to open")
}

func run(cmd *cobra.Command, args []string) {
	broker := NewBroker(address,
		userAuthHandler,
		senderHandler,
		electron.SASLAllowedMechs("ANONYMOUS"),
		electron.SASLAllowInsecure(true))
	if err := broker.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
func senderHandler(s electron.Sender) <-chan amqp.Message {
	ch := make(chan amqp.Message, 1)
	go func(s electron.Sender) {
		for i := 0; i < 5; i++ {
			// TODO: Get Message from MQ
			topic := s.Source()
			ch <- amqp.NewMessageWith(topic)
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
