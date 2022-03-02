package mq

import "os"

var (
	_Brokers       = "http://"
	_ConsumerGroup = "amqp"
)

func Init() {
	if urlEnv := os.Getenv("KAFKA_URL"); urlEnv != "" {
		_Brokers = urlEnv
	}

	if groupEnv := os.Getenv("KAFKA_CONSUMER_GROUP"); groupEnv != "" {
		_ConsumerGroup = groupEnv
	}
}
