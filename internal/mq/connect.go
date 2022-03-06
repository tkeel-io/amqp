package mq

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

func Connect(ctx context.Context, topics string) {
	keepRunning := true
	fmt.Println("Starting a new Sarama consumer")

	version, err := sarama.ParseKafkaVersion(_KafkaVersion)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version

	switch _KafkaAssignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", _KafkaAssignor)
	}

	if _KafkaOldestEable {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
		ctx:   ctx,
	}

	client, err := sarama.NewConsumerGroup(strings.Split(_BrokersStr, ","), _ConsumerGroup, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	consumptionIsPaused := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Debug("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			fmt.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			fmt.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Info("Resuming consumption")
	} else {
		client.PauseAll()
		log.Info("Pausing consumption")
	}

	*isPaused = !*isPaused
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ctx   context.Context
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	id, ok := consumer.ctx.Value("id").(string)
	if !ok {
		return errors.New("id not found in context")
	}
	ShutdownChan(id)
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	id, ok := consumer.ctx.Value("id").(string)
	if !ok {
		return errors.New("id not found in context")
	}
	ch := FindSourceChan(id)

	for message := range claim.Messages() {
		fmt.Printf("Message claimed: value = %s, timestamp = %v, topic = %s \n", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
		data, err := parser(message.Value)
		if err != nil {
			log.Errorf("Error parsing message: %v", err)
			return err
		}
		select {
		case ch <- data:
			log.Info("get data from subscription kafka topic: ", data)
		}
	}

	return nil
}

var channelManager = make(map[string]chan []byte)

func FindSourceChan(id string) chan []byte {
	ch, ok := channelManager[id]
	if ok {
		return ch
	}
	ch = make(chan []byte, 1)
	channelManager[id] = ch
	return ch
}

func ShutdownChan(id string) {
	ch, ok := channelManager[id]
	if ok {
		close(ch)
		delete(channelManager, id)
	}
}
