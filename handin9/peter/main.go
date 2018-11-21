package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
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
			time.Sleep(1 * time.Second)
			return client
		}
	}

	// If we got down here, it's a new network
	fmt.Println("[Creating a new network for the client]")
	network := &Network{}
	network.Initialize(client, ip)
	networks = append(networks, network)
	time.Sleep(1 * time.Second)

	return client
}

func sendTransaction(from *Client, to *Client, amount int) bool {
	if from.ownPeer == to.ownPeer {
		fmt.Println("from and to cannot be the same")
	} else if amount < 1 {
		fmt.Println("Cannot make a transaction of less than 1 AU")
	} else {

		id := from.ownPeer.Address + "-" + strconv.Itoa(from.transactionID)
		from.transactionID++

		transaction := SignedTransaction{ID: id, From: from.ownPeer.Pk, To: to.ownPeer.Pk, Amount: amount}

		messageForSigning := GenerateMessageFromTransaction(&transaction)

		signature := Sign(messageForSigning, from.sk)
		transaction.Signature = signature.String()

		if transaction.isValid() {
			var message = Message{ID: TRANSACTION_MESSAGE, Value: transaction}

			//fmt.Println("Sending transaction with id ", id, " from ", from.ownPeer.Address, " to ", to.ownPeer.Address, " for ", amount, "Msg:", message)

			from.handleTransaction(message)

			return true
		}
	}

	return false
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

func benchmarkTransactions(network *Network) {

	numClients := math.Min(10, float64(len(network.Clients)))

	if numClients < 2 {
		fmt.Println("Cannot run benchmark on a network with less than 2 peers")
		printArrow()
		return
	}

	NUM_TRANSACTIONS := 1000
	start := time.Now()
	sent := 0
	// Send 1000 transactions
	for i := 0; i < NUM_TRANSACTIONS; i++ {
		r1 := int(math.Floor(rand.Float64() * float64(numClients)))
		r2 := r1
		for r2 == r1 {
			r2 = int(math.Floor(rand.Float64() * float64(numClients)))
		}

		from := network.Clients[r1]
		to := network.Clients[r2]

		didSend := sendTransaction(from, to, 2)

		if didSend {
			sent++
		}
	}

	fmt.Println("Sent", sent, "/", NUM_TRANSACTIONS, "transactions successfully")

	// Wait until every client has gotten all the transactions
	for i := 0; i < len(network.Clients); i++ {
		client := network.Clients[i]

		// We only need 90% of the transactions to arrive, because of a bug that makes valid transactions invalid
		for len(client.transactionsReceived) < sent {
			time.Sleep(1 * time.Millisecond)
		}

		fmt.Println("Client", i, "has ", len(client.transactionsReceived), "transactions")
	}

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("Time elapsed:", elapsed)
	printArrow()
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

	InitConsts()

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

	} else if cmCheck("calc", 0) {

		fmt.Println("Calculating the average Val for a 90% threshold")

		var edges []*big.Int
		for j := 0; j < 100; j++ {
			fmt.Println("Round", j)
			pair := KeyGen(2000)
			var vals []*big.Int
			var blockID = 0
			for i := 0; i < 1000; i++ {
				drawMsgStr := strconv.Itoa(123 + blockID)
				drawMsg := []byte(drawMsgStr)
				draw := Sign(drawMsg, pair.Sk)

				shaMsg := []byte(drawMsgStr + pair.Pk.toString() + draw.String())
				sha := sha256.New()
				sha.Write(shaMsg)
				hash := sha.Sum(nil)

				val := new(big.Int).Mul(big.NewInt(int64(PREMIUM_ACCOUNT)), new(big.Int).SetBytes(hash))

				vals = append(vals, val)
				blockID++
			}

			bigIntSort := func(i, j int) bool {
				a := vals[i]
				b := vals[j]
				res := a.Cmp(b)
				return res >= 0
			}
			sort.SliceStable(vals, bigIntSort)

			index := int(math.Floor(0.1 * float64(len(vals))))
			edges = append(edges, vals[index])
		}

		sum := big.NewInt(0)

		for _, edge := range edges {
			sum.Add(sum, edge)
		}

		avg := new(big.Int).Div(sum, big.NewInt(int64(len(edges))))

		fmt.Println("Average:", avg)

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
			fmt.Println("##### Network", i, "#####")
			network := networks[i]

			numClients := len(network.Clients)
			var first *Ledger
			var matches = 0

			for j := 0; j < numClients; j++ {
				fmt.Println("Client", j)
				ledger, _ := network.Clients[i].generateNewestLedger()
				if ledger != nil {
					ledger.PrintStatus()
					fmt.Println()

					if first == nil {
						first = ledger
						matches = 1
					} else if first.Match(ledger) {
						matches++
					}
				}
			}

			fmt.Println("Got", matches, "/", numClients, "matches")
		}
	} else if cmCheck("list ls", 0) {
		for i := 0; i < len(networks); i++ {

			network := networks[i]
			length := len(network.Clients)

			fmt.Println("Network", i, "contains", length, "client(s)")

			for j := 0; j < length; j++ {
				client := network.Clients[j]
				fmt.Println("Client", j, "is connected to", len(client.peers), "peers, and has ip", client.ownPeer.Address, " and key:", client.ownPeer.Pk[0:50]+"...")
			}

			fmt.Println()
		}
	} else if cmCheck("benchmark bm", 1) {
		index, err := strconv.Atoi(params[0])

		checkError(err, "Invalid index")

		if gotError {
			return
		}

		network := networks[index]

		fmt.Println("Please wait until the benchmark completes")
		fmt.Println("Also note, that this command should only be called before any transactions have been made")

		go benchmarkTransactions(network)

	} else if cmCheck("keys", 0) {
		fmt.Println("Listing all King keys in each network\n")
		for networkIndex, network := range networks {

			fmt.Println("Network", networkIndex)

			for keyIndex, pair := range network.KingKeys {
				fmt.Println("Public key", keyIndex, ":", pair.Pk.toString())
			}

			fmt.Println()
		}
	} else if cmCheck("help", 0) {
		fmt.Println("Displaying a list of all commands:\n")

		fmt.Println("createClient | cc\t<ip : string>")
		fmt.Println("Creates a new client and adds it to an exsisting network, if the IP matches another peer\n")

		fmt.Println("setup\t<numClients : int>")
		fmt.Println("Setup a number of clients in a network, that will connect to each other randomly\n")

		fmt.Println("trans\t<network : int> <from index : int> <to index : int> <amount : int>")
		fmt.Println("Makes a transaction between two clients. Use \"list\" to see all clients in the network alongside their index\n")

		fmt.Println("calc")
		fmt.Println("Calculates the average Val for a 90% threshold. This is used to calculate an estimated hardness\n")

		fmt.Println("start\t<network : int>")
		fmt.Println("Begins running the lottery for a network of clients. Do not call this function more than once per network\n")

		fmt.Println("status")
		fmt.Println("Goes through each network and prints the ledger for each client, alongside how many ledgers match\n")

		fmt.Println("list | ls\t")
		fmt.Println("Lists all the peers in each network\n")

		fmt.Println("benchmark | bm\t<network : int>")
		fmt.Println("Runs a benchmark on a network, where 1000 transactions are sent randomly between all peers. This will display the time it takse for all the transactions to arrive\n")

		fmt.Println("keys\t")
		fmt.Println("Lists all the king keys for each network\n")

		fmt.Println("help\t")
		fmt.Println("Lists this list of commands\n")

		fmt.Println("quit")
		fmt.Println("Exits the program")

	} else if cmCheck("setup", 1) {
		numClients, err := strconv.Atoi(params[0])

		if err != nil {
			fmt.Println("Invalid number of clients")
			return
		}

		var cs []*Client

		for i := 0; i < numClients; i++ {
			var client *Client
			if len(cs) > 0 {
				index := int(math.Floor(rand.Float64() * float64(len(cs))))
				prevClient := cs[index]
				client = createClient(prevClient.ownPeer.Address)
			} else {
				client = createClient("")
			}

			cs = append(cs, client)
		}

	} else if cmCheck("test", 1) {

		handleCommand("setup " + params[0])

		numClients, _ := strconv.Atoi(params[0])

		if numClients < 2 {
			fmt.Println("Num clients needs to be at least 2")
			numClients = 2
		}

		handleCommand("trans 0 0 1 100")
		handleCommand("trans 0 0 1 100")
		handleCommand("trans 0 1 0 1000")
		handleCommand("trans 0 0 1 50")
		handleCommand("trans 0 0 1 0")   // Invalid
		handleCommand("trans 0 0 1 -69") // Invalid
		handleCommand("trans 0 0 1 10000000")
		handleCommand("start 0")
	} else if cmCheck("quit", 0) {
		fmt.Println("Thanks for playing")
		terminate = true
		return
	} else if !invalidCommand {
		fmt.Println("Invalid command. Type \"help\" for a list of commands")
	}
}
