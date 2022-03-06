package service

import "os"

var (
	_subscribeValidateURL = "http://192.168.123.9:30707/apis/core-broker/v1/validate/subscribe"
)

func ConfigInit() {
	s := os.Getenv("CORE_BROKER_SUBSCRIBE_VALIDATE_URL")
	if s != "" {
		_subscribeValidateURL = s
	}
}
