package bus

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// mutex for devices m

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}

// MsgCtrl is an internal structure to pack messages together with
// info about sender and receivers
type MsgCtrl struct {
	sender   *Node
	payload  *Message
	receiver string
}

// Bus provides a mechanism for the send  messages to a
// collection of channels in unicast/broadcast way
type Bus struct {
	in       chan MsgCtrl
	close    chan bool
	waitsync chan bool
	nodes    []*Node
	nodeLock sync.Mutex
}

// NewBus creates a new broadcast bus.
func NewBus() *Bus {
	in := make(chan MsgCtrl)
	close := make(chan bool)
	waitsync := make(chan bool)
	return &Bus{in: in, close: close, waitsync: waitsync}
}

// Destroy returns the number of nodes in the Broadcast Bus.
func (b *Bus) Destroy() {
	b.Stop()
	close(b.in)
	close(b.close)
	close(b.waitsync)
}

// NodeCount returns the number of nodes in the Broadcast Bus.
func (b *Bus) NodeCount() int {
	return len(b.Nodes())
}

// Nodes returns a slice of Nodes that are currently in the Bus.
func (b *Bus) Nodes() []*Node {
	b.nodeLock.Lock()
	res := b.nodes[:]
	b.nodeLock.Unlock()
	return res
}

// Join  handles the attachment to the bus
func (b *Bus) Join(n *Node) {
	n.AttachToBus(b)
	b.nodeLock.Lock()
	b.nodes = append(b.nodes, n)
	b.nodeLock.Unlock()
}

// Leave removes the provided node from the bus
func (b *Bus) Leave(leaving *Node) error {
	b.nodeLock.Lock()
	defer b.nodeLock.Unlock()
	nodeIndex := -1
	for index, node := range b.nodes {
		if node == leaving {
			nodeIndex = index
			break
		}
	}
	if nodeIndex == -1 {
		return errors.New("Could not find provided member for removal")
	}
	b.nodes = append(b.nodes[:nodeIndex], b.nodes[nodeIndex+1:]...)
	// leaving.close <- true

	return nil
}

// Stop terminates the bus process immediately.
func (b *Bus) Stop() {
	b.close <- true
}

// Start checks for some message in the broadcast queue pending to send
func (b *Bus) Start() {
	for {
		select {
		case received := <-b.in:
			nodes := make([]*Node, len(b.nodes))
			switch received.receiver {
			case "all":
				b.nodeLock.Lock()
				copy(nodes, b.nodes)
				b.nodeLock.Unlock()
			default:
				id := received.receiver
				b.nodeLock.Lock()
				for _, n := range b.nodes {
					if n.ID == id {
						nodes = append(nodes, n)
					}
				}
				b.nodeLock.Unlock()
			}

			log.Info("BUS: begin message distribution to nodes")
			var wg sync.WaitGroup
			for _, node := range nodes {
				// This is done in a goroutine because if it
				// weren't it would be a blocking call
				wg.Add(1)
				go func(node *Node, received MsgCtrl) {
					defer wg.Done()
					node.Read <- received.payload
				}(node, received)
			}
			wg.Wait()
			log.Info("BUS: End send message distribution to nodes")
			b.waitsync <- true
			log.Debug("BUS: sync sent")

		case <-b.close:
			return
		}
	}
}

// Send send message to one receiver to the Bus
func (b *Bus) Send(id string, m *Message) {
	b.in <- MsgCtrl{sender: nil, payload: m, receiver: id}
	log.Debugf("BUS: unicast message %s , %+v sent to node %s", m.Type.String(), m.Data, id)
	<-b.waitsync
}

// Broadcast send message to all nodes attached to the bus
func (b *Bus) Broadcast(m *Message) {
	b.in <- MsgCtrl{sender: nil, payload: m, receiver: "all"}
	log.Debugf("BUS: Broadcast message %s , %+v sent", m.Type.String(), m.Data)
	<-b.waitsync
}
