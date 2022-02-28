package main

import (
	"fmt"
	"os"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/spf13/cobra"
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

// TODO: Authenticate users here
func userAuthHandler(electron.Connection) (interface{}, error) {
	return nil, nil
}

// senderHandler process with message witch u want to send
func senderHandler(s electron.Sender) <-chan amqp.Message {
	ch := make(chan amqp.Message, 1)
	go func(s electron.Sender) {
		for i := 0; i < 5; i++ {
			// TODO: Get Message from MQ
			ch <- amqp.NewMessageWith(s.Source())
		}
	}(s)
	return ch
}
