package amqp

import (
	"fmt"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReceive(t *testing.T) {
	go func() {
		testUrl := "amqp://localhost:3172/topic"
		r, err := NewReceiver(testUrl, electron.User("fred1"),
			electron.VirtualHost("Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NjE0MDI5Miwic3ViIjoidXNyLTIzNDgwMmM5YWQwY2NjOGUxYTViYWQ0NWZiNmMifQ.n3xo5lavvWz5tBV-Gs0UPFafP69Aumfn2L38DTm_E_VVUhLG7SblTBZqgtlyjHfD5qVmJH8iIsJmy-hkAWYz4w"),
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
	assert.Error(t, err)
	for i := 0; i < 5; i++ {
		content, err := r.Receive()
		assert.NoError(t, err)
		fmt.Printf("Received %d: %s\n", i, content)
	}
	time.Sleep(3 * time.Second)
}
