package amqp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/tkeel-io/kit/log"
)

type Broker struct {
	address       string                // broker address
	container     electron.Container    // electron.Container manages AMQP connections.
	outcome       chan electron.Outcome // Channel to receive the Outcome of sent messages.
	opts          []electron.ConnectionOption
	authHandler   func(electron.Connection, electron.Sender) (interface{}, error)
	senderHandler func(context.Context, electron.Sender) <-chan amqp.Message
}

func NewBroker(address string,
	authHandler func(electron.Connection, electron.Sender) (interface{}, error),
	senderHandler func(context.Context, electron.Sender) <-chan amqp.Message,
	opts ...electron.ConnectionOption) *Broker {
	b := &Broker{
		address:       address,
		container:     electron.NewContainer("server"),
		outcome:       make(chan electron.Outcome, 1),
		opts:          opts,
		authHandler:   authHandler,
		senderHandler: senderHandler,
	}

	return b
}

// Run listens for incoming net.Conn connections and starts an electron.Connection for each one.
func (b *Broker) Run() error {
	listener, err := net.Listen("tcp", b.address)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Printf("Listening on %v\n", listener.Addr())

	// Start a goroutine for each new connections
	for {
		c, err := b.container.Accept(listener, b.opts...)
		if err != nil {
			log.Debugf("Error accepting connectionAccept error: %v", err)
			continue
		}
		ctx, cancel := context.WithCancel(context.Background())
		cc := &connection{b, c, nil, ctx, cancel}
		go cc.run() // Handle the conn
		log.Debugf("Accepted %v", c)
	}
}

// State for a broker connection
type connection struct {
	broker *Broker
	conn   electron.Connection
	auth   interface{}
	ctx    context.Context
	cancel context.CancelFunc
}

// accept remotely-opened endpoints (Session, SenderManager and Receiver) on a connection
// and start goroutines to service them.
func (c *connection) run() {
	for in := range c.conn.Incoming() {
		log.Debugf("incoming %v", in)

		switch in := in.(type) {
		case *electron.IncomingSender:
			s := in.Accept().(electron.Sender)
			go c.sender(s)
		case *electron.IncomingSession, *electron.IncomingConnection:
			fmt.Println("connect")
			in.Accept()
		case *electron.IncomingReceiver:
			in.SetPrefetch(true)
			in.SetCapacity(100)
			r := in.Accept().(electron.Receiver)
			go c.receiver(r)

		default:
			in.Accept() // Accept sessions unconditionally
		}
	}
	fmt.Printf("incoming closed: %v\n", c.conn)
	c.cancel()
}

// receiver receives messages and pushes to a queue.
func (c *connection) receiver(receiver electron.Receiver) {
	q := make(chan amqp.Message, 1)
	for {
		if rm, err := receiver.Receive(); err == nil {
			log.Debugf("%v: received %v", receiver, rm.Message.Body())
			q <- rm.Message
			rm.Accept()
		} else {
			log.Debugf("%v error: %v", receiver, err)
			break
		}
	}
}

func (c *connection) sender(sender electron.Sender) {
	c.conn.Sync()
	auth, err := c.broker.authHandler(c.conn, sender)
	if err != nil {
		log.Error(err)
		fmt.Printf("auth error: %v\n", err)
		c.conn.Disconnect(err)
		return
	}
	c.auth = auth

	ch := c.broker.senderHandler(c.ctx, sender)
	for {
		if sender.Error() != nil {
			log.Debugf("%v closed: %v", sender, sender.Error())
			fmt.Printf("%v closed: %v\n", sender, sender.Error())
			return
		}
		select {
		case m, ok := <-ch:
			if ok {
				log.Debugf("%v: sent %v", sender, m.Body())
				fmt.Printf("%v: sent %v\n", sender, m.Body())
				sender.SendSync(m)
			} else {
				log.Debugf("%v: channel closed", sender)
				fmt.Printf("%v: channel closed\n", sender)
				sender.Close(errors.New("sender closed"))
			}
		case <-sender.Done():
			break
		}
	}
}
