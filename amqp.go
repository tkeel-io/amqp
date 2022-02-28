package amqp

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/apache/qpid-proton/go/pkg/amqp"
	"github.com/apache/qpid-proton/go/pkg/electron"
	"github.com/tkeel-io/kit/log"
)

type Broker struct {
	address   string                // broker address
	queues    queues                // A collection of queues.
	container electron.Container    // electron.Container manages AMQP connections.
	sent      chan sentMessage      // Channel to record sent messages.
	acks      chan electron.Outcome // Channel to receive the Outcome of sent messages.
}

func NewBroker(address string) *Broker {
	qsize := 100
	b := &Broker{
		address:   address,
		queues:    makeQueues(qsize),
		container: electron.NewContainer(fmt.Sprintf("broker[%v]", os.Getpid())),
		acks:      make(chan electron.Outcome),
		sent:      make(chan sentMessage),
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

	go b.acknowledgements() // Handles acknowledgements for all connections.

	// Start a goroutine for each new connections
	for {
		c, err := b.container.Accept(listener)
		if err != nil {
			log.Debugf("Error accepting connectionAccept error: %v", err)
			continue
		}
		cc := &connection{b, c}
		go cc.run() // Handle the connection
		log.Debugf("Accepted %v", c)
	}
}

// State for a broker connection
type connection struct {
	broker     *Broker
	connection electron.Connection
}

// accept remotely-opened endpoints (Session, SenderManager and Receiver) on a connection
// and start goroutines to service them.
func (c *connection) run() {
	for in := range c.connection.Incoming() {
		log.Debugf("incoming %v", in)

		switch in := in.(type) {

		case *electron.IncomingSender:
			s := in.Accept().(electron.Sender)
			go c.sender(s)

		case *electron.IncomingReceiver:
			in.SetPrefetch(true)
			in.SetCapacity(100) // Pre-fetch up to credit window.
			r := in.Accept().(electron.Receiver)
			go c.receiver(r)

		default:
			in.Accept() // Accept sessions unconditionally
		}
	}
	log.Debugf("incoming closed: %v", c.connection)
}

// receiver receives messages and pushes to a queue.
func (c *connection) receiver(receiver electron.Receiver) {
	q := c.broker.queues.Get(receiver.Target())
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

// sender pops messages from a queue and sends them.
func (c *connection) sender(sender electron.Sender) {
	q := c.broker.queues.Get(sender.Source())
	for {
		if sender.Error() != nil {
			log.Debugf("%v closed: %v", sender, sender.Error())
			return
		}

		select {
		case m := <-q:
			log.Debugf("%v: sent %v", sender, m.Body())
			sm := sentMessage{m, q}
			c.broker.sent <- sm                    // Record sent message
			sender.SendAsync(m, c.broker.acks, sm) // Receive outcome on c.broker.acks with Value sm

		case <-sender.Done(): // break if sender is closed
			break
		}
	}
}

// acknowledgements keeps track of sent messages and receives outcomes.
//
// We could have handled outcomes separately per-connection, per-sender or even
// per-message. Message outcomes are returned via channels defined by the user
// so they can be grouped in any way that suits the application.
func (b *Broker) acknowledgements() {
	sentMap := make(map[sentMessage]bool)
	for {
		select {
		case sm, ok := <-b.sent: // A local sender records that it has sent a message.
			if ok {
				sentMap[sm] = true
			} else {
				return // Closed
			}
		case outcome := <-b.acks: // The message outcome is available
			sm := outcome.Value.(sentMessage)
			delete(sentMap, sm)
			if outcome.Status != electron.Accepted { // Error, release or rejection
				sm.q.PutBack(sm.m) // Put the message back on the queue.
			}
		}
	}
}

// Record of a sent message and the queue it came from.
// If a message is rejected or not acknowledged due to a failure, we will put it back on the queue.
type sentMessage struct {
	m amqp.Message
	q queue
}

// Use a buffered channel as a very simple queue.
type queue chan amqp.Message

// PutBack Put a message back on the queue, does not block.
func (q queue) PutBack(m amqp.Message) {
	select {
	case q <- m:
	default:
		// Not an efficient implementation but ensures we don't block the caller.
		go func() { q <- m }()
	}
}

// Concurrent-safe map of queues.
type queues struct {
	queueSize int
	m         map[string]queue
	lock      sync.Mutex
}

func makeQueues(queueSize int) queues {
	return queues{queueSize: queueSize, m: make(map[string]queue)}
}

// Get Create a queue if not found.
func (qs *queues) Get(name string) queue {
	qs.lock.Lock()
	defer qs.lock.Unlock()
	q := qs.m[name]
	if q == nil {
		q = make(queue, qs.queueSize)
		qs.m[name] = q
	}
	return q
}
