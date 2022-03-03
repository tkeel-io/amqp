package amqp

import (
	"fmt"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReceive(t *testing.T) {
	testUrl := "amqp://localhost:3172/IyH99QyiW35HwICm"
	r, err := NewReceiver(testUrl, electron.User("fred1"),
		electron.VirtualHost("Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NjI5NTgzNSwic3ViIjoidXNyLTIzNDgwMmM5YWQwY2NjOGUxYTViYWQ0NWZiNmMifQ.SMRoyGQ6hGzugRoqRN_n69kEhx6A4PWoCUrYZYsNKQpdNb4bTNl8ct3YAMbVLhIO7HN1dIJd0stgwh6OKaH2Tg"),
		electron.Password([]byte("mypassword")),
		electron.SASLAllowInsecure(true))
	assert.NoError(t, err)
	for {
		content, err := r.Receive()
		assert.NoError(t, err)

		fmt.Printf("Received: %v\n", content)
	}

}
