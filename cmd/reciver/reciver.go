package main

import (
	"fmt"
	amqplib "github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tkeel-io/amqp"
	"github.com/tkeel-io/amqp/internal/jsonutil"
	"os"
	"strings"
)

var (
	connectionAddr string
	token          string
	count          int
)

func main() {
	cmd := &cobra.Command{
		Use:   "amqp-receiver",
		Short: "amqp receiver tool, help you to receive message from amqp connectionAddr",
		Long:  "amqp receiver tool, help you to receive message from amqp connectionAddr.",
		Run:   receiverService,
	}

	cmd.PersistentFlags().StringVarP(&connectionAddr, "connect", "c", "amqp://localhost:5672", "amqp connectionAddr url which you want to connect")
	cmd.PersistentFlags().StringVarP(&token, "token", "t", "", "token flag for you set your Authorization token quickly")
	cmd.PersistentFlags().IntVar(&count, "count", 0, "receive message count, default is 0, means receive all message until CTRL + C interrupt.")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(0)
}

func receiverService(cmd *cobra.Command, args []string) {
	connectionAddr = validateAMQPAddr(connectionAddr)
	color.Green("Try to Connect %s \n", connectionAddr)
	r, err := amqp.NewReceiver(connectionAddr)
	if err != nil {
		color.Red("Connect %s failed, err: %s \n", connectionAddr, err)
		return
	}
	if token != "" {
		r, err = amqp.NewReceiver(connectionAddr, setTokenOpts(token)...)
		if err != nil {
			color.Red("Connect %s failed, err: %s \n", connectionAddr, err)
			return
		}
	}

	// receive message util CTRL + C interrupt
	if count == 0 {
		for {
			if err = receiveAndPrint(r); err != nil {
				return
			}
		}
	}

	// receive message util count
	for i := 0; i < count; i++ {
		if err = receiveAndPrint(r); err != nil {
			return
		}
	}
}

func validateAMQPAddr(addr string) string {
	if strings.HasPrefix(addr, "amqp://") {
		return addr
	}
	return "amqp://" + addr
}

func setTokenOpts(token string) []electron.ConnectionOption {
	if !strings.HasPrefix(token, "Bearer") {
		token = "Bearer " + token
	}
	opts := []electron.ConnectionOption{
		electron.SASLAllowInsecure(true),
		electron.VirtualHost(token),
	}

	color.Yellow("Set token: %s \n", token)

	return opts
}

func receiveAndPrint(r *amqp.Receiver) error {
	content, err := r.Receive()
	if err != nil {
		color.Red("Receive failed, err: %s \n", err)
		return err
	}
	color.Green("[RECEIVED RAW DATA]")
	fmt.Printf("%s \n", jsonutil.Stringify([]byte(content.(amqplib.Binary))))
	return nil
}
