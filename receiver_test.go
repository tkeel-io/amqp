package main

import (
	"fmt"
	"testing"

	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/stretchr/testify/assert"
)

func TestReceive(t *testing.T) {
	go func() {
		testUrl := "amqp://localhost:3172/topic"
		r, err := NewReceiver(testUrl, electron.User("fred1"),
			electron.VirtualHost("token"),
			electron.Password([]byte("mypassword")),
			electron.SASLAllowInsecure(true))
		assert.NoError(t, err)
		for i := 0; i < 5; i++ {
			content, err := r.Receive()
			assert.NoError(t, err)
			fmt.Printf("Received %d: %s\n", i, content)
		}
	}()
	testUrl := "amqp://localhost:3172/topic2"
	r, err := NewReceiver(testUrl, electron.User("fred1"),
		electron.VirtualHost("token"),
		electron.Password([]byte("mypassword")),
		electron.SASLAllowInsecure(true))
	assert.NoError(t, err)
	for i := 0; i < 5; i++ {
		content, err := r.Receive()
		assert.NoError(t, err)
		fmt.Printf("Received %d: %s\n", i, content)
	}
}
