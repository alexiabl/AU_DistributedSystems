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
	conn, err := net.Dial("tcp", targetIP)

	if err != nil {
		fmt.Println("No peer found or invalid IP/Port")
		c.firstPeer = true
	} else {
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

		c.peers = initInfo.Peers
		c.setGenesisBlock(&initInfo.GenesisBlock)
	}
}

func (c *Client) sendMessage(conn net.Conn, msg Message) {
	var enc = gob.NewEncoder(conn)
	var err = enc.Encode(&msg)
	if err != nil {
		fmt.Println("Got error when sending message: ", err.Error())
		fmt.Println("Message:", msg)
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
				c.handleTransaction(message)
				break

			case NEW_PEER_MESSAGE:
				var peer = message.Value.(Peer)

				// If the peer is already registered
				if c.isPeerRegistered(peer.Address) {
					break
				}

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

func (c *Client) handleTransaction(msg Message) {
	var transaction = msg.Value.(SignedTransaction)
	var transID = transaction.ID

	// Don't broadcast an invalid message
	if !transaction.isValid() {
		fmt.Println("Received an invalid transaction")
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
	c.transactionsReceived = append(c.transactionsReceived, transaction)

	c.outboundMessages <- msg
}

func (c *Client) handleBlock(block *Block, msg Message) {

	// Skip this block, if it has already been received
	for i := 0; i < len(c.blocks); i++ {
		if c.blocks[i].ID == block.ID && c.blocks[i].Sender == block.Sender {
			return
		}
	}

	senderPk := GeneratePublicKeyFromString(block.Sender)

	if c.IsValidDraw(c.genesisBlock.Seed, block.ID, block.Draw, senderPk) {
		if block.isValid() {
			if c.isBlockValid(block) {

				c.blocks = append(c.blocks, block)
				c.outboundMessages <- msg
				return
			}
		} else {
			fmt.Println("Invalid block: unable to match the signature with the block")
		}
	}

	fmt.Println("[Warning] Received an invalid block")
}

func (c *Client) isBlockValid(block *Block) bool {

	prev := c.getBlockBySignature(block.PreviousBlock)

	if prev != nil {
		if block.ID > prev.ID { // Verify that prev block is smaller than current
			if block.ID <= c.currentBlockID+1 { // Verify that the block is the expected one
				return true
			} else {
				fmt.Println("Invalid block: block ID is too far ahead. Got", block.ID, "but is currently on", c.currentBlockID)
			}
		} else {
			fmt.Println("Invalid block: the block ID is lower than that of the previous block")
		}
	} else {
		fmt.Println("Invalid block: unable to locate the previous block")
	}

	return false
}

func (c *Client) getBlockBySignature(sign string) *Block {
	for i := 0; i < len(c.blocks); i++ {

		if c.blocks[i].Signature == sign {
			return c.blocks[i]
		}
	}

	return nil
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

	prevSign := c.getLongestBlock(c.currentBlockID).Signature
	draw := GenerateDraw(c.genesisBlock.Seed, c.currentBlockID, c.sk)

	block := &Block{c.currentBlockID, prevSign, c.ownPeer.Pk, transactions, "", draw}
	c.SignBlock(block)

	return block
}

func (c *Client) SignBlock(block *Block) {
	blockMsg := GenerateMessageFromBlock(block)
	signature := Sign(blockMsg, c.sk).String()
	block.Signature = signature
}

func (c *Client) setGenesisBlock(genesis *GenesisBlock) {
	if !genesis.isValid() {
		panic("Got an invalid genesis key")
	}

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

/* Returns a ledger alongside a boolean describing if the ledger is valid or not.
The ledger will be invalid only if a block brings an account below 0. */
func (c *Client) generateNewestLedger() (*Ledger, bool) {
	block := c.getLongestBlock(MAX_INT)

	return c.generateLedgerForBlock(block)
}

func (c *Client) generateLedgerForBlock(block *Block) (*Ledger, bool) {
	// Make a copy of the current ledger
	ledger := MakeLedger()

	blocks := []*Block{block}

	for block.PreviousBlock != "" {
		block = c.getBlockBySignature(block.PreviousBlock)

		if block == nil {
			fmt.Println("Unable to generate ledger: missing block with ID", block.PreviousBlock[0:50])
			return nil, true
		}

		blocks = append([]*Block{block}, blocks...) // Unshift the block
	}

	for _, kingKey := range c.genesisBlock.KingKeys {
		ledger.AddAmount(kingKey, PREMIUM_ACCOUNT)
	}

	fmt.Println("There are ", len(blocks), "blocks")

	var usedTransactions []string

	for _, block := range blocks {
		senderPay := 10 // Add 10 AU to the sender of the block
		for _, transID := range block.Transactions {

			// Skip this ID, if it has already been seen
			isUsed := false
			for _, id := range usedTransactions {
				if id == transID {
					isUsed = true
					break
				}
			}
			if isUsed {
				continue
			}

			foundTrans := false
			for _, trans := range c.transactionsReceived {
				if transID == trans.ID {
					foundTrans = true
					usedTransactions = append(usedTransactions, transID)
					if ledger.SignedTransaction(&trans) {
						senderPay++ // Add 1 AU for each valid transaction
						break
					} else {
						fmt.Println("Transaction: " + trans.ID + " is invalid")
					}
				}
			}

			if !foundTrans {
				fmt.Println("Missing a transaction:", transID+". Cannot calculate ledger")
				return nil, true
			}
		}

		senderPeer := c.GetPeerFromPK(block.Sender)
		fmt.Println("Adding", senderPay, "to", senderPeer.Address)
		ledger.AddAmount(block.Sender, senderPay)
	}

	return ledger, true
}

func (c *Client) startBlocks() {
	// Only start the block if the cliens has a King key
	for _, key := range c.genesisBlock.KingKeys {
		if key == c.ownPeer.Pk {
			go c.blockTimer()
			return
		}
	}
}

func (c *Client) blockTimer() {
	ticker := time.NewTicker(SLOT_LENGTH)

	for {
		select {
		case <-ticker.C:
			c.currentBlockID++

			draw := GenerateDraw(c.genesisBlock.Seed, c.currentBlockID, c.sk)
			val := c.CalculateDrawValue(c.genesisBlock.Seed, c.currentBlockID, draw, c.pk)

			if val.Cmp(HARDNESS) < 0 {
				continue
			}

			fmt.Println("Got a valid draw from", c.ownPeer.Address, "with val", val.String())
			printArrow()

			block := c.generateBlock()
			block.Draw = draw
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
