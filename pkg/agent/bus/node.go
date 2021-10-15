package bus

// Command define valid message types to be passed using the bus
type Command int

// TODO documentar cada comando
const (
	// Exit without waiting for anything
	Exit Command = iota
	// SyncExit order the device and its measurement goroutines to exit and waits till them are finished
	SyncExit
	Enabled
	LogLevel
	ForceGather
	FilterUpdate
	SNMPResetHard
	SNMPReset
	SNMPDebug
	SetSNMPMaxRep
)

// Message a basic message type
type Message struct {
	Type Command
	Data interface{}
}

// Node represents node of a Broadcast bus.
type Node struct {
	ID   string
	bus  *Bus
	Read chan *Message
}

// NewNode create a newNode unattached struc
func NewNode(id string) *Node {
	nodeChannel := make(chan *Message)
	return &Node{
		ID:   id,
		Read: nodeChannel,
	}
}

// AttachToBus set the bus to work
func (n *Node) AttachToBus(b *Bus) {
	n.bus = b
}

// Close removes the node it is called on from its broadcast bus.
func (n *Node) Close() {
	log.Debugf("Closing node %s...", n.ID)
	n.bus.Leave(n)
	close(n.Read)
}

// SendMsg  a message to itself one Node to the channels of all
// the other nodes in its bus.
func (n *Node) SendMsg(m *Message) {
	n.Read <- m
}

// Broadcast  a message to all other nodes of the bus.
func (n *Node) Broadcast(m *Message) {
	n.bus.in <- MsgCtrl{sender: n, payload: m, receiver: "all"}
}

// RecvMsg reads one value from the node's Read channel
func (n *Node) RecvMsg() *Message {
	return <-n.Read
}
