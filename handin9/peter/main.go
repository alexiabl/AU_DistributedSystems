package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var networks = []*Network{}

func createClient(ip string) *Client {

	client := &Client{}

	// Check if the client connects to an already existing network
	for i := 0; i < len(networks); i++ {
		network := networks[i]

		if network.ContainsClientWithIP(ip) {
			network.AddClient(client, ip)

			fmt.Println("[Added client to existing network]")
			return client
		}
	}

	// If we got down here, it's a new network
	fmt.Println("[Creating a new network for the client]")
	network := &Network{}
	network.Initialize(client, ip)
	networks = append(networks, network)

	return client
}

func sendTransaction(from *Client, to *Client, amount int) {
	if from.ownPeer == to.ownPeer {
		fmt.Println("from and to cannot be the same")
	} else {
		id := from.ownPeer.Address + "-" + strconv.Itoa(from.transactionID)
		from.transactionID++

		transaction := SignedTransaction{ID: id, From: from.ownPeer.Pk, To: to.ownPeer.Pk, Amount: amount}

		messageForSigning := GenerateMessageFromTransaction(&transaction)

		signature := Sign(messageForSigning, from.sk)
		transaction.Signature = signature.String()

		if transaction.isValid() {
			var message = Message{ID: TRANSACTION_MESSAGE, Value: transaction}

			fmt.Println("Sending transaction with id ", id, " from ", from, " to ", to, " for ", amount)

			from.outboundMessages <- message
		}
	}
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

func printArrow() {
	fmt.Print("> ")
}

func checkError(err error, msg string) {
	if err != nil {
		gotError = true
		fmt.Println(msg)
	}
}

func checkRange(index int, len int) {
	if index < 0 || index > len-1 {
		gotError = true
		fmt.Println("Index", index, "is out of range 0 -", len-1)
	}
}

var gotError = false
var terminate = false

var fullIP = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}:\d{1,5}$`)
var onlyPort = regexp.MustCompile(`^:\d{1,5}$`)

func main() {

	// Register gob interfaces - need for en-/decoding
	gob.Register(Peer{})
	gob.Register([]Peer{})
	gob.Register(SignedTransaction{})
	gob.Register(Block{})
	gob.Register(GenesisBlock{})
	gob.Register(InitInfo{})

	// Start listening for user input
	reader := bufio.NewReader(os.Stdin)
	for {
		printArrow()

		text, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		handleCommand(text)

		if terminate {
			return
		}
	}
	/*// Start listening for input
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
					// Adding "(you)" after the local ip in the list
					youStr := ""
					if peer == ownPeer {
						youStr = "(you)"
					}

					fmt.Println(strconv.Itoa(index+1)+": "+peer.Address, youStr)
				}
			}

			// Print the account of each peer
			if text == "status\n" {
				ledger.PrintStatus()
			}

			// List all possible commands
			if text == "help\n" {
				fmt.Println("All available commands are: quit, list, status, help, trans, testSignature")
			}

			if text == "testSignature\n" {
				fmt.Println("Sending a signed transaction with an invalid signature")

				if len(peers) < 2 {
					fmt.Println("Not enough peers in the network")
					continue
				}

				var sender = ownPeer
				var receiver Peer

				for i := 0; i < len(peers); i++ {
					if peers[i] != ownPeer {
						receiver = peers[i]
						break
					}
				}

				amount := 123
				id := sender.Address + "-" + strconv.Itoa(transactionID)
				transactionID++

				transaction := SignedTransaction{ID: id, From: sender.Pk, To: receiver.Pk, Amount: amount}

				// Switched transaction.From and transaction.To, which will give another message and therefore a differnet signature, than it is supposed to be
				messageForSigning := []byte(transaction.To + transaction.From + id + string(amount))

				signature := Sign(messageForSigning, sk)
				transaction.Signature = signature.String()

				if handleTransaction(transaction) {
					fmt.Println("The transaction was accepted as real")
				} else {
					fmt.Println("The transaction wasn't accepted")
				}
			}

			// Make a transaction
			var splitMessage = strings.Split(text, " ")
			if splitMessage[0] == "trans" {

				if len(splitMessage) != 3 {
					fmt.Println("Use:\ntrans <to IP> <amount>")
				} else {
					var from = ownPeer.Address
					var to = splitMessage[1]

					if from == to {
						fmt.Println("<to IP> needs to be someone else than yourself")
					} else if !isPeerRegistered(to) {
						fmt.Println("<to IP> not found in peers")
					} else {
						sender := ownPeer
						receiver := GetPeerFromIP(to)
						amountStr := strings.Replace(splitMessage[2], "\n", "", -1)
						amount, _ := strconv.Atoi(amountStr)
						id := sender.Address + "-" + strconv.Itoa(transactionID)
						transactionID++

						transaction := SignedTransaction{ID: id, From: sender.Pk, To: receiver.Pk, Amount: amount}

						messageForSigning := GenerateMessageFromTransaction(&transaction)

						signature := Sign(messageForSigning, sk)
						transaction.Signature = signature.String()

						if handleTransaction(transaction) {
							var message = Message{ID: TRANSACTION_MESSAGE, Value: transaction}

							fmt.Println("Sending transaction with id ", id, " from ", from, " to ", to, " for ", amount)

							outboundMessages <- message
						}
					}
				}
			}

			printArrow()
		}
	}*/
}

func handleCommand(text string) {
	gotError = false

	commandParts := strings.Split(strings.Trim(text, "\n"), " ")
	command := strings.ToLower(commandParts[0])
	params := commandParts[1:]

	invalidCommand := false
	cmCheck := func(exptectedString string, expectedArgs int) bool {
		exptectedStrings := strings.Split(exptectedString, " ")

		for i := 0; i < len(exptectedStrings); i++ {
			str := exptectedStrings[i]
			if strings.ToLower(str) == command {
				if expectedArgs == len(params) {
					return true
				}

				invalidCommand = true
				fmt.Println("Invalid number of arguments. Expected", expectedArgs, "but got", len(params))
			}
		}

		return false
	}

	if cmCheck("createClient cc", 1) {
		fmt.Println("Creating a new client")

		ip := params[0]

		if onlyPort.MatchString(ip) {
			ip = getOwnAddress() + ip
		} else if !fullIP.MatchString(ip) {
			ip = "0.0.0.0:00000"
			fmt.Println("This is not a valid ip. Using", ip)
		}

		createClient(ip)
		fmt.Println("Finished creating client")

	} else if cmCheck("trans", 4) {
		networkIndex, errNetwork := strconv.Atoi(params[0])
		fromIndex, errFrom := strconv.Atoi(params[1])
		toIndex, errTo := strconv.Atoi(params[2])
		amount, errAmount := strconv.Atoi(params[3])

		checkError(errNetwork, "Invalid network index")
		checkError(errFrom, "Invalid from id")
		checkError(errTo, "Invalid to id")
		checkError(errAmount, "Invalid amount")

		checkRange(networkIndex, len(networks))

		if gotError {
			return
		}

		network := networks[networkIndex]

		checkRange(fromIndex, len(network.Clients))
		checkRange(toIndex, len(network.Clients))

		if gotError {
			return
		}

		from := network.Clients[fromIndex]
		to := network.Clients[toIndex]

		sendTransaction(from, to, amount)

	} else if cmCheck("start", 1) {
		index, err := strconv.Atoi(params[0])

		checkError(err, "Invalid index")

		if gotError {
			return
		}

		network := networks[index]
		for i := 0; i < len(network.Clients); i++ {
			network.Clients[i].startBlocks()
		}

	} else if cmCheck("status", 0) {
		for i := 0; i < len(networks); i++ {
			fmt.Println("##### Network", i, ":")
			network := networks[i]
			for j := 0; j < len(network.Clients); j++ {
				fmt.Println("Client", j)
				network.Clients[i].generateNewestLedger().PrintStatus()
				fmt.Println()
			}
		}
	} else if cmCheck("networks", 0) {
		// Print all the networks along side how many is in each
	} else if cmCheck("list ls", 0) {
		for i := 0; i < len(networks); i++ {

			network := networks[i]
			len := len(network.Clients)

			fmt.Println("Network", i, "contains", len, "client(s)")

			for j := 0; j < len; j++ {
				client := network.Clients[j]
				fmt.Println("Client", j, "has ip", client.ownPeer.Address)
			}

			fmt.Println()
		}
	} else if cmCheck("keys", 1) {
		// Print list of king keys in a given network
	} else if cmCheck("help", 0) {
		fmt.Println("A list of commands:")
		fmt.Println("\trans\t<network : int> <from index : int> <to index : int> <amount : int>")
		fmt.Println("\tcreateClient | cc\t<ip : string> <RSA key index : int>")

	} else if cmCheck("test", 0) {
		c := createClient("")
		createClient(c.ownPeer.Address)

		handleCommand("trans 0 0 1 4567")
		handleCommand("start 0")
	} else if cmCheck("quit", 0) {
		fmt.Println("Thanks for playing")
		terminate = true
		return
	} else if !invalidCommand {
		fmt.Println("Invalid command. Type \"help\" for a list of commands")
	}
}
