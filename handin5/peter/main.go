package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	connections = append(connections, conn)

	otherEnd := conn.RemoteAddr().String()

	for {
		var decoder = gob.NewDecoder(conn)
		var message = Message{}
		var err = decoder.Decode(&message)

		if err != nil {
			fmt.Println("Ending session with ", otherEnd)
			printArrow()

			// Finding the index of conn in connections
			index := -1

			for connIndex, tempConn := range connections {
				if conn == tempConn {
					index = connIndex
					break
				}
			}

			// Remove the connection from the array
			if index != -1 {
				connections = append(connections[:index], connections[index+1:]...)
			}

			return
		} else {

			switch message.ID {
			case TRANSACTION_MESSAGE:
				var transaction = message.Value.(SignedTransaction)
				var transID = transaction.From + transaction.ID

				// If this transaction has already been sent, break
				for i := 0; i < len(transactionsSent); i++ {
					if transactionsSent[i] == transID {
						break
					}
				}

				handleTransaction(transaction)

				fmt.Println("[Got transaction]")
				printArrow()

				outboundMessages <- message
				break
			case NEW_PEER_MESSAGE:
				var peer = message.Value.(Peer)

				// If the peer is already registered
				if isPeerRegistered(peer.Address) {
					break
				}

				peers = append(peers, peer)
				ledger.InitializeAccount(peer)
				sortPeers()

				outboundMessages <- message
				break
			case REQUEST_PEER_LIST_MESSAGE:
				var response = Message{ID: PEER_LIST_MESSAGE, Value: peers}
				sendMessage(conn, response)
				break
			}
		}
	}
}

func listenForConnections(ln net.Listener) {
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a new connection ", conn.RemoteAddr())
		printArrow()
		go handleConnection(conn)
	}
}

func broadcastMessages() {
	for {
		var message = <-outboundMessages

		for i := 0; i < len(connections); i++ {
			var conn = connections[i]
			sendMessage(conn, message)
		}
	}
}

func sendMessage(conn net.Conn, msg Message) {
	var enc = gob.NewEncoder(conn)
	var err = enc.Encode(&msg)
	if err != nil {
		fmt.Println("Got error when sending message: ", err.Error())
	}
}

func isPeerRegistered(address string) bool {
	for i := 0; i < len(peers); i++ {
		if peers[i].Address == address {
			return true
		}
	}

	return false
}

func printArrow() {
	fmt.Print("> ")
}

func getOwnAddress() string {
	var hostName, _ = os.Hostname()
	var addrs, _ = net.LookupIP(hostName)
	var ipv4 net.IP
	for _, addr := range addrs {
		if ipv4 = addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	return "127.0.0.1"
}

func getPeerList() {
	reader := bufio.NewReader(os.Stdin)

	// Get IP
	fmt.Print("Enter IP address> ")
	ip, _ := reader.ReadString('\n')

	if ip == "" {
		ip = getOwnAddress()
	}

	// Get port
	fmt.Print("Enter port> ")
	port, _ := reader.ReadString('\n')

	// Try to establish connection
	fullAddress := strings.Replace(ip+":"+port, "\n", "", -1)
	conn, err := net.Dial("tcp", fullAddress)
	if err != nil {
		fmt.Println("No peer found or invalid IP/Port")
		fmt.Println("Starting a new network")
	} else {
		fmt.Println("Connection successful")

		defer conn.Close()

		//Request peer list
		var message = Message{ID: REQUEST_PEER_LIST_MESSAGE}
		sendMessage(conn, message)

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
		peers = list
	}
}

func connectToPeers() {

	fmt.Println("Connecting to up to 10 peers in the network")

	var len = len(peers)
	var index = -1
	for i := 0; i < len; i++ {
		peer := peers[i]
		if peer == ownPeer {
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
		peer := peers[currentIndex]

		// If the list is exhausted
		if peer == ownPeer {
			return
		}

		// Connect to the peer
		conn, err := net.Dial("tcp", peer.Address)
		if err != nil {
			fmt.Println("Unable to connect to peer: ", peer.Address)
		} else {
			fmt.Println("Connected to: ", peer.Address)

			go handleConnection(conn)
		}
	}
}

func registerPeersInLedger() {
	for i := 0; i < len(peers); i++ {
		ledger.InitializeAccount(peers[i])
	}
}

func handleTransaction(trans SignedTransaction) {
	ledger.SignedTransaction(&trans)
	var transID = trans.From + trans.ID
	transactionsSent = append(transactionsSent, transID)
}

func broadcastSelf() {
	var message = Message{ID: NEW_PEER_MESSAGE, Value: ownPeer}
	outboundMessages <- message
}

func setupListeningServer() net.Listener {
	fmt.Println("Listening for connections on:")
	ln, _ := net.Listen("tcp", ":")

	// Printing the IP's
	var address = getOwnAddress()

	// Printing the port
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	// Generate address
	fullAddress := address + ":" + port
	fmt.Println("Listening on address:", fullAddress)

	ownPeer = Peer{Address: fullAddress, Pk: pk.toString()}

	printArrow()

	return ln
}

func addSelfToList() {
	peers = append(peers, ownPeer)
	sortPeers()
}

func sortPeers() {
	address := func(i, j int) bool {
		return peers[i].Address < peers[j].Address
	}
	sort.SliceStable(peers, address)
}

var outboundMessages = make(chan Message) // A channel for all messages
var connections = []net.Conn{}            // A list of all current connections
var transactionsSent = []string{}         // A list of all received messages
var peers = []Peer{}                      // List of all peers in the network
var ownPeer Peer                          // The id of this peer (public key as string)

var pk PublicKey
var sk SecretKey

var transactionID = 0
var ledger *Ledger

func main() {

	// Generate RSA key
	n, d := KeyGen(2000)
	pk = generatePublicKey(n, e)
	sk = generateSecretKey(n, d)

	// Register gob interfaces - need for en-/decoding
	gob.Register(Peer{})
	gob.Register(SignedTransaction{})

	// Creates the ledger
	ledger = MakeLedger()

	// Connect to a peer in the network, and get the list of peers
	getPeerList()

	// Start listening for new connections
	var ln = setupListeningServer()
	go listenForConnections(ln)

	// Start broadcasting messagesages
	go broadcastMessages()

	addSelfToList()
	registerPeersInLedger()
	connectToPeers()
	broadcastSelf()

	// Ready
	fmt.Println("[Ready]")
	printArrow()

	// Start listening for input
	for {
		reader := bufio.NewReader(os.Stdin)

		// Begin chat loop
		for {
			text, _ := reader.ReadString('\n')

			// Exit the program if the user types 'quit'
			if text == "quit\n" {
				return
			}

			// List all peers in the peer list
			if text == "list\n" {
				for index, peer := range peers {
					fmt.Println(strconv.Itoa(index+1) + ": " + peer.Address)
				}
			}

			// Print the account of each peer
			if text == "status\n" {
				ledger.PrintStatus()
			}

			// List all possible commands
			if text == "help\n" {
				fmt.Println("All available commands are: quit, list, status, help, trans")
			}

			// Make a transaction
			var splitMessage = strings.Split(text, " ")
			if splitMessage[0] == "trans" {

				if len(splitMessage) != 4 {
					fmt.Println("Use:\ntrans <from> <to> <amount>")
				} else {
					var from = splitMessage[1]
					var to = splitMessage[2]

					if from == to {
						fmt.Println("From and to cannot be the same")
					} else if !isPeerRegistered(from) && !isPeerRegistered(to) {
						fmt.Println("<from> or <to> not found in peers")
					} else {
						var amountStr = strings.Replace(splitMessage[3], "\n", "", -1)
						var amount, _ = strconv.Atoi(amountStr)
						var id = pk.toString() + "-" + strconv.Itoa(transactionID)
						transactionID++
						var transaction = SignedTransaction{ID: id, From: from, To: to, Amount: amount}
						var message = Message{ID: TRANSACTION_MESSAGE, Value: transaction}

						handleTransaction(transaction)

						fmt.Println("Sending transaction with id ", id, " from ", from, " to ", to, " for ", amount)

						outboundMessages <- message
					}
				}
			}

			printArrow()
		}
	}
}
