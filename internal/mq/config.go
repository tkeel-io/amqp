package mq

import (
	"os"
)

var (
	_BrokersStr       = "localhost:9092"
	_ConsumerGroup    = "amqp"
	_KafkaVersion     = "3.1.0"
	_KafkaAssignor    = "roundrobin"
	_KafkaOldestEable = true
)

var ConsumedData = make(chan []byte, 1)

func Init() {
	if urlEnv := os.Getenv("KAFKA_BROKERS"); urlEnv != "" {
		_BrokersStr = urlEnv
	}

	if groupEnv := os.Getenv("KAFKA_CONSUMER_GROUP"); groupEnv != "" {
		_ConsumerGroup = groupEnv
	}

	if versionEnv := os.Getenv("KAFKA_VERSION"); versionEnv != "" {
		_KafkaVersion = versionEnv
	}

	if assignorEnv := os.Getenv("KAFKA_ASSIGNOR"); assignorEnv != "" {
		_KafkaAssignor = assignorEnv
	}

	if oldestEnv := os.Getenv("KAFKA_OLDEST_ENABLE"); oldestEnv != "" {
		_KafkaOldestEable = oldestEnv == "true"
	}

}
