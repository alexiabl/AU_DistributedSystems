package main

type Network struct {
	Clients []Client
	Ledger  *Ledger
}

func (n *Network) Initialize(initClient Client) {
	n.Clients = []Client{initClient}
	n.Ledger = initClient.ledger
}

func (n *Network) AddClient(client Client) {
	n.Clients = append(n.Clients, client)
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
