package amqp

import (
	"strings"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
)

type Receiver struct {
	r         electron.Receiver
	Url       string
	ConnOpts  []electron.ConnectionOption
	Container electron.Container
	Conn      electron.Connection
}

func NewReceiver(urlStr string, opts ...electron.ConnectionOption) (*Receiver, error) {
	r := &Receiver{
		Container: electron.NewContainer("receiver-" + urlStr),
		Url:       urlStr,
		ConnOpts:  opts,
	}
	if err := r.run(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Receiver) Receive() (interface{}, error) {
	msg, err := r.r.Receive()
	if err != nil {
		return nil, err
	}

	msg.Accept()
	return msg.Message.Body(), nil
}

func (r *Receiver) Close() {
	if r.Conn != nil {
		r.Conn.Close(nil)
	}
	if r.r != nil {
		r.r.Close(nil)
	}
}

func (r *Receiver) run() error {
	url, err := amqp.ParseURL(r.Url)
	if err != nil {
		return err
	}
	c, err := electron.Dial("tcp", url.Host, r.ConnOpts...) // NOTE: Dial takes just the Host part of the URL
	if err != nil {
		return err
	}
	addr := strings.TrimPrefix(url.Path, "/")
	recv, err := c.Receiver(electron.Source(addr))
	if err != nil {
		return err
	}

	r.Conn = c
	r.r = recv
	return nil
}
