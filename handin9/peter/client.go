package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"sort"
	"time"
)

type Client struct {
	outboundMessages     chan Message        // A channel for all messages
	connections          []net.Conn          // A list of all current connections
	transactionsSent     []string            // A list of already broadcasted transactions
	transactionsReceived []SignedTransaction // A list of all received transactions
	peers                []Peer              // List of all peers in the network
	ownPeer              Peer                // The id of this peer (public key as string)
	firstPeer            bool                // Indicated if this client is the first peer in the network
	blocks               []*Block            // A list of all received blocks
	genesisBlock         *GenesisBlock
	currentBlockID       int

	pk PublicKey
	sk SecretKey

	transactionID int
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
		var message = Message{ID: REQUEST_INIT_INFO_MESSAGE}
		c.sendMessage(conn, message)

		// Wait for response
		var newMessage = &Message{}
		var dec = gob.NewDecoder(conn)
		var err = dec.Decode(newMessage)
		if err != nil {
			fmt.Println("Error while reading init info: ", err.Error())
			return
		} else if newMessage.ID != INIT_INFO_MESSAGE {
			fmt.Println("Got an unexspected response from other peer: " + newMessage.ID)
			return
		}

		initInfo, ok := newMessage.Value.(InitInfo)

		if !ok {
			panic("Error while decoding InitInfo from message")
		}

		fmt.Println("[Got initial info]")
		c.peers = initInfo.Peers
		c.setGenesisBlock(&initInfo.GenesisBlock)

		fmt.Println("Genesis block:", c.genesisBlock.KingKeys)
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
				c.handleTransaction(transaction, message)
				break

			case NEW_PEER_MESSAGE:
				var peer = message.Value.(Peer)

				fmt.Println("got peer message:", peer)

				// If the peer is already registered
				if c.isPeerRegistered(peer.Address) {
					break
				}

				fmt.Println("Wasn't already added")

				c.peers = append(c.peers, peer)
				c.sortPeers()

				c.outboundMessages <- message
				break

			case REQUEST_INIT_INFO_MESSAGE:
				initInfo := InitInfo{Peers: c.peers, GenesisBlock: *c.genesisBlock}
				response := Message{ID: INIT_INFO_MESSAGE, Value: initInfo}
				c.sendMessage(conn, response)
				break

			case BLOCK_MESSAGE:
				block := message.Value.(Block)
				c.handleBlock(&block, message)
				break
			}
		}
	}
}

func (c *Client) handleTransaction(trans SignedTransaction, msg Message) {
	var transID = trans.ID

	// Don't broadcast an invalid message
	if !trans.isValid() {
		return
	}

	// If this transaction has already been sent, break
	alreadySent := false
	for i := 0; i < len(c.transactionsSent); i++ {
		if c.transactionsSent[i] == transID {
			alreadySent = true
			break
		}
	}
	if alreadySent {
		return
	}

	c.transactionsSent = append(c.transactionsSent, transID)
	c.transactionsReceived = append(c.transactionsReceived, trans)

	fmt.Println("[Got transaction]")
	printArrow()

	c.outboundMessages <- msg
}

func (c *Client) handleBlock(block *Block, msg Message) {

	for i := 0; i < len(c.blocks); i++ {
		if c.blocks[i].ID == block.ID && c.blocks[i].Sender == block.Sender {
			return
		}
	}

	fmt.Println("Me:", c.ownPeer.Address)
	fmt.Println("Peers:", c.peers)
	if block.isValidSignature() {
		if c.isBlockValid(block) {
			if len(block.Transactions) > 0 {
				fmt.Println("Got a valid block with", len(block.Transactions), "transactions")
			}

			c.blocks = append(c.blocks, block)

			c.outboundMessages <- msg
			/*
				currentBlockIndex = block.ID
				for i := 0; i<len(block.Transactions); i++ {
					trans := block.Transactions[i]
					for j:= 0; j<len(transactionsReceived); j++ {
						if trans == transactionsReceived[j].ID {
							currTransaction := transactionsReceived[j]
							if (handleTransaction(currTransaction)){
								fmt.Println("Success!")
								printArrow()
								break
							}else{
								fmt.Println("Transaction: "+currTransaction.ID+" failed")
							}

						}
					}
				}
			*/
		}
	}
}

func (c *Client) isBlockValid(block *Block) bool {

	prev := c.getBlockBySignature(block.PreviousBlock)

	if block.ID > prev.ID { // Verify that prev block is smaller than current
		if block.ID <= c.currentBlockID+1 { // Verify that the block is the expected one
			return true
		}
	}

	return false
}

func (c *Client) getBlockBySignature(sign string) *Block {
	for i := 0; i < len(c.blocks); i++ {

		if c.blocks[i].Signature == sign {
			return c.blocks[i]
		}
	}

	return c.genesisBlock.Block
}

func (c *Client) getLongestBlock(lessThanID int) *Block {

	getDistanceToGenesis := func(block *Block) int {
		var dist = 0

		for block.ID > 0 {
			dist++
			block = c.getBlockBySignature(block.PreviousBlock)
		}

		return dist
	}

	var longestDist = 0
	var longestBlock = c.genesisBlock.Block

	for i := 0; i < len(c.blocks); i++ {
		block := c.blocks[i]

		if block.ID >= lessThanID {
			continue
		}

		dist := getDistanceToGenesis(block)

		if dist > longestDist {
			longestDist = dist
			longestBlock = block
		} else if dist == longestDist {

			// If distance is the same, pick the one with the biggest ID
			if block.ID == longestBlock.ID {
				if block.Signature > longestBlock.Signature {
					longestDist = dist
					longestBlock = block
				}
			} else if block.ID > longestBlock.ID {
				longestDist = dist
				longestBlock = block
			}
		}
	}

	return longestBlock
}

func (c *Client) generateBlock() *Block {

	// Copy the list of transactions
	transactions := make([]string, len(c.transactionsReceived))

	for i := 0; i < len(c.transactionsReceived); i++ {
		transactions[i] = c.transactionsReceived[i].ID
	}

	getTransIndex := func(transID string) int {
		for i := 0; i < len(transactions); i++ {
			if transactions[i] == transID {
				return i
			}
		}

		return -1
	}

	// Remove all 'old' transactions
	for i := 0; i < len(c.blocks); i++ {
		block := c.blocks[i]
		for j := 0; j < len(block.Transactions); j++ {
			index := getTransIndex(block.Transactions[j])
			if index != -1 {
				transactions = append(transactions[:index], transactions[index+1:]...)
			}
		}
	}

	prevSign := ""
	if c.genesisBlock != nil {
		prevSign = c.getLongestBlock(c.currentBlockID).Signature
	}

	block := &Block{c.currentBlockID, prevSign, c.ownPeer.Pk, transactions, ""}
	blockMsg := GenerateMessageFromBlock(block)
	signature := Sign(blockMsg, c.sk).String()
	block.Signature = signature

	return block
}

func (c *Client) setGenesisBlock(genesis *GenesisBlock) {
	//TODO Verify genesis block
	c.genesisBlock = genesis
	c.blocks = append(c.blocks, genesis.Block)
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

func (c *Client) broadcastSelf() {
	var message = Message{ID: NEW_PEER_MESSAGE, Value: c.ownPeer}
	c.outboundMessages <- message
}

func (c *Client) addSelfToList() {
	c.peers = append(c.peers, c.ownPeer)
	c.sortPeers()
}

func (c *Client) generateNewestLedger() *Ledger {
	// Make a copy of the current ledger
	ledger := MakeLedger()

	block := c.getLongestBlock(MAX_INT)
	blocks := []*Block{}

	for block.ID != 0 {
		fmt.Println("ID:", block.ID)
		blocks = append([]*Block{block}, blocks...) // Unshift the block
		block = c.getBlockBySignature(block.PreviousBlock)
	}

	for _, kingKey := range c.genesisBlock.KingKeys {
		ledger.InitializePremiumAccount(kingKey)
	}

	for _, block := range blocks {
		for _, transID := range block.Transactions {
			for _, trans := range c.transactionsReceived {
				if transID == trans.ID {
					if ledger.SignedTransaction(&trans) {
						fmt.Println("Success!")
						printArrow()
						break
					} else {
						fmt.Println("Transaction: " + trans.ID + " failed")
					}

				}
			}
		}
	}

	return ledger
}

func (c *Client) startBlocks() {
	go c.blockTimer()
}

func (c *Client) blockTimer() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			c.currentBlockID++

			block := c.generateBlock()
			msg := Message{ID: BLOCK_MESSAGE, Value: block}
			c.handleBlock(block, msg)
		}
	}
}

func (c *Client) Initialize(targetIP string, pair KeyPair) {

	c.pk = pair.Pk
	c.sk = pair.Sk

	c.outboundMessages = make(chan Message)

	// Connect to a peer in the network, and get the list of peers
	c.getPeerList(targetIP)

	// Start listening for new connections
	var ln = c.setupListeningServer()
	go c.listenForConnections(ln)

	// Start broadcasting messagesages
	go c.broadcastMessages()

	c.addSelfToList()

	if !c.firstPeer || true {
		c.connectToPeers()
		c.broadcastSelf()
	}
}
