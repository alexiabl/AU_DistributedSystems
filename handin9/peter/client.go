package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"sort"
)

type Client struct {
	outboundMessages chan Message // A channel for all messages
	connections      []net.Conn   // A list of all current connections
	transactionsSent []string     // A list of all received messages
	peers            []Peer       // List of all peers in the network
	ownPeer          Peer         // The id of this peer (public key as string)
	firstPeer        bool         // Indicated if this client is the first peer in the network

	pk PublicKey
	sk SecretKey

	transactionID int
	ledger        *Ledger
}

func (c *Client) GetPeerFromPK(str string) *Peer {
	for i := 0; i < len(c.peers); i++ {
		if str == c.peers[i].Pk {
			return &c.peers[i]
		}
	}

	return nil
}

func (c *Client) GetPeerFromIP(ip string) *Peer {
	for i := 0; i < len(c.peers); i++ {
		if ip == c.peers[i].Address {
			return &c.peers[i]
		}
	}

	return nil
}

func (c *Client) getPeerList(targetIP string) {

	// Try to establish connection
	fmt.Println("Trying to connect to other peer...")
	conn, err := net.Dial("tcp", targetIP)

	if err != nil {
		fmt.Println("No peer found or invalid IP/Port")
		c.firstPeer = true
	} else {
		fmt.Println("Connection successful")
		c.firstPeer = false

		defer conn.Close()

		//Request peer list
		var message = Message{ID: REQUEST_PEER_LIST_MESSAGE}
		c.sendMessage(conn, message)

		// Wait for response
		var newMessage = &Message{}
		var dec = gob.NewDecoder(conn)
		var err = dec.Decode(newMessage)
		if err != nil {
			fmt.Println("Error while reading peer list: ", err.Error())
			return
		} else if newMessage.ID != PEER_LIST_MESSAGE {
			fmt.Println("Got an unexspected response from other peer: " + newMessage.ID)
			return
		}

		var list = newMessage.Value.([]Peer)
		fmt.Println("[Got list of peers]")
		c.peers = list
	}
}

func (c *Client) sendMessage(conn net.Conn, msg Message) {
	var enc = gob.NewEncoder(conn)
	var err = enc.Encode(&msg)
	if err != nil {
		fmt.Println("Got error when sending message: ", err.Error())
	}
}

func (c *Client) setupListeningServer() net.Listener {
	fmt.Println("Listening for connections on:")
	ln, _ := net.Listen("tcp", ":")

	// Printing the IP's
	var address = getOwnAddress()

	// Printing the port
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	// Generate address
	fullAddress := address + ":" + port
	fmt.Println(fullAddress)

	c.ownPeer = Peer{Address: fullAddress, Pk: c.pk.toString()}

	return ln
}

func (c *Client) listenForConnections(ln net.Listener) {
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a new connection ", conn.RemoteAddr())
		go c.handleConnection(conn)
	}
}

func (c *Client) handleConnection(conn net.Conn) {
	defer conn.Close()

	c.connections = append(c.connections, conn)

	for {
		var decoder = gob.NewDecoder(conn)
		var message = Message{}
		var err = decoder.Decode(&message)

		if err != nil {

			// Finding the index of conn in connections
			index := -1

			for connIndex, tempConn := range c.connections {
				if conn == tempConn {
					index = connIndex
					break
				}
			}

			// Remove the connection from the array
			if index != -1 {
				c.connections = append(c.connections[:index], c.connections[index+1:]...)
			}

			return
		} else {

			switch message.ID {
			case TRANSACTION_MESSAGE:
				var transaction = message.Value.(SignedTransaction)
				var transID = transaction.ID

				// If this transaction has already been sent, break
				alreadySent := false
				for i := 0; i < len(c.transactionsSent); i++ {
					if c.transactionsSent[i] == transID {
						alreadySent = true
						break
					}
				}
				if alreadySent {
					break
				}

				if !c.handleTransaction(transaction) {
					break
				}

				fmt.Println("[Got transaction]")
				printArrow()

				c.outboundMessages <- message
				break
			case NEW_PEER_MESSAGE:
				var peer = message.Value.(Peer)

				// If the peer is already registered
				if c.isPeerRegistered(peer.Address) {
					break
				}

				c.peers = append(c.peers, peer)
				c.ledger.InitializeAccount(peer)
				c.sortPeers()

				c.outboundMessages <- message
				break
			case REQUEST_PEER_LIST_MESSAGE:
				var response = Message{ID: PEER_LIST_MESSAGE, Value: c.peers}
				c.sendMessage(conn, response)
				break
			}
		}
	}
}

func (c *Client) isPeerRegistered(address string) bool {
	for i := 0; i < len(c.peers); i++ {
		if c.peers[i].Address == address {
			return true
		}
	}

	return false
}

func (c *Client) sortPeers() {
	address := func(i, j int) bool {
		return c.peers[i].Address < c.peers[j].Address
	}
	sort.SliceStable(c.peers, address)
}

func (c *Client) handleTransaction(trans SignedTransaction) bool {
	valid := c.ledger.SignedTransaction(&trans)

	if !valid {
		return false
	}

	c.transactionsSent = append(c.transactionsSent, trans.ID)
	return true
}

func (c *Client) broadcastMessages() {
	for {
		var message = <-c.outboundMessages

		for i := 0; i < len(c.connections); i++ {
			var conn = c.connections[i]
			c.sendMessage(conn, message)
		}
	}
}

func (c *Client) connectToPeers() {

	fmt.Println("Connecting to up to 10 peers in the network")

	// Find the index of itself
	var len = len(c.peers)
	var index = -1
	for i := 0; i < len; i++ {
		peer := c.peers[i]
		if peer == c.ownPeer {
			index = i
			break
		}
	}

	if index == -1 {
		fmt.Println("Error: peer ID wasn't in the list of peers")
		return
	}

	// Connect to the 10 peers after peerID in the list with wrap around
	for i := 1; i <= 10; i++ {
		currentIndex := (index + i) % len
		peer := c.peers[currentIndex]

		// If the list is exhausted
		if peer == c.ownPeer {
			return
		}

		// Connect to the peer
		conn, err := net.Dial("tcp", peer.Address)
		if err != nil {
			fmt.Println("Unable to connect to peer: ", peer.Address)
		} else {
			fmt.Println("Connected to: ", peer.Address)

			go c.handleConnection(conn)
		}
	}
}

func (c *Client) registerPeersInLedger() {
	for i := 0; i < len(c.peers); i++ {
		c.ledger.InitializeAccount(c.peers[i])
	}
}

func (c *Client) broadcastSelf() {
	var message = Message{ID: NEW_PEER_MESSAGE, Value: c.ownPeer}
	c.outboundMessages <- message
}

func (c *Client) addSelfToList() {
	c.peers = append(c.peers, c.ownPeer)
	c.sortPeers()
}

func (c *Client) Initialize(targetIP string, pk *PublicKey, sk *SecretKey) {

	if pk == nil || sk == nil {
		pk, sk := KeyGen(2000)
		c.pk = pk
		c.sk = sk
	} else {
		c.pk = *pk
		c.sk = *sk
	}

	c.outboundMessages = make(chan Message)

	// Creates the ledger
	c.ledger = MakeLedger()

	// Connect to a peer in the network, and get the list of peers
	c.getPeerList(targetIP)

	// Start listening for new connections
	var ln = c.setupListeningServer()
	go c.listenForConnections(ln)

	// Start broadcasting messagesages
	go c.broadcastMessages()

	c.addSelfToList()
	c.registerPeersInLedger()

	if !c.firstPeer || true {
		c.connectToPeers()
		c.broadcastSelf()
	}
}
