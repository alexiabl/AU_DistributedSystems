package main

import (
	"fmt"
	"reflect"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	Lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

type SignedTransaction struct {
	ID        string // Any string
	From      string // A verification key coded as a string
	To        string // A verification key coded as a string
	Amount    int    // Amount to transfer
	Signature string // Potential signature coded as string
}

type Peer struct {
	Address string
	Pk      string
}

type Message struct {
	ID    string
	Value interface{}
}

const TRANSACTION_MESSAGE = "transMsg"                 // When the message contains a transaction
const NEW_PEER_MESSAGE = "newPeerMsg"                  // When the message contains a new peer
const REQUEST_PEER_LIST_MESSAGE = "requestPeerListMsg" // When the message requests the list of peers
const PEER_LIST_MESSAGE = "peerListMsg"                // When the message contains the list of peers

func (l *Ledger) InitializeAccount(peer Peer) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.Accounts[peer.Pk] = 0
}

func (l *Ledger) PrintStatus() {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	var keys = reflect.ValueOf(l.Accounts).MapKeys()
	fmt.Println("There are", len(keys), "keys")
	for i := 0; i < len(keys); i++ {
		var key = keys[i]
		var str = key.String()
		fmt.Println("Account", i, "has", l.Accounts[str], "dineros")
	}
}
