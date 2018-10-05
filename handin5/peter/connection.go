package main

type ConnectionManager interface {
	GetMessageChannel() chan Message
	StartListening() string                        // Start listening for incoming connections and return the address
	SendMessage(m Message)                         // Sends a message to all connected peers
	SendSpecificMessage(m Message, address string) // Sends a message specifically to one peer
	ConnectTo(address string)                      // Connects to a peer
	DisconnectFrom(address string)                 // Disconnects from a peer
}
