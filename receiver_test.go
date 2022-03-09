package amqp

import (
	"fmt"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReceive(t *testing.T) {
	testUrl := "amqp://localhost:5672"
	r, err := NewReceiver(testUrl,
		electron.VirtualHost("Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NjMxMTA0OCwic3ViIjoidXNyLTIzNDgwMmM5YWQwY2NjOGUxYTViYWQ0NWZiNmMifQ.mv4k9uzoSAY_l876xNysOL_wkFpp4o1GemuEEBBh_fuRuDi0naDEzon5ET-_o60Y4KSXDXd4QuZi5rhzzuDxhw"),
		electron.SASLAllowInsecure(true))
	assert.NoError(t, err)
	for {
		content, err := r.Receive()
		assert.NoError(t, err)

		fmt.Printf("Received: %v\n", content)
	}

}
