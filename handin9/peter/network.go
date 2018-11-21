package main

type Network struct {
	Clients  []*Client
	KingKeys []KeyPair
	KeyIndex int
}

func (n *Network) Initialize(initClient *Client, ip string) {
	// Generate King Keys
	publicKingKeys := make([]string, 10) // Contains all the keys as strings
	n.KingKeys = make([]KeyPair, 10)

	for i := 0; i < 10; i++ {
		pair := KeyGen(2000)
		n.KingKeys[i] = pair
		publicKingKeys[i] = pair.Pk.toString()
	}

	n.KeyIndex = 0

	// Save the client
	n.Clients = []*Client{}

	n.AddClient(initClient, ip)

	// Generate genesis block
	block := &Block{0, "", initClient.ownPeer.Pk, []string{}, "", GenerateDraw(SEED, 0, initClient.sk)}
	initClient.SignBlock(block)
	genesisBlock := GenesisBlock{block, publicKingKeys, SEED}
	initClient.setGenesisBlock(&genesisBlock)
}

func (n *Network) GetNextKey() KeyPair {
	if n.KeyIndex >= len(n.KingKeys) {
		return KeyGen(2000)
	} else {
		pair := n.KingKeys[n.KeyIndex]
		n.KeyIndex++
		return pair
	}
}

func (n *Network) AddClient(client *Client, ip string) {
	n.Clients = append(n.Clients, client)
	client.Initialize(ip, n.GetNextKey())
}

func (n *Network) ContainsClientWithIP(ip string) bool {
	len := len(n.Clients)
	for i := 0; i < len; i++ {
		client := n.Clients[i]

		if client.ownPeer.Address == ip {
			return true
		}
	}

	return false
}
