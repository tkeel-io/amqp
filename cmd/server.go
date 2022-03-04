package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tkeel-io/amqp/internal/service"
)

var address string

func ParseFlags(root *cobra.Command) {
	root.PersistentFlags().StringVar(&address, "address", ":3172", "the broker address and port you want to open")
}

var Server = &cobra.Command{
	Use:   "amqp-broker",
	Short: "amqp broker starter",
	Long:  "amqp broker starter",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	service.ConfigInit()
	broker := service.NewBroker(address)
	if err := broker.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
