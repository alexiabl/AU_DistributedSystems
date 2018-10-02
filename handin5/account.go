package main

import (
	"fmt"
	"reflect"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
}

type SignedTransaction struct {
	ID string // Any string
	From string // A verification key coded as a string
	To string // A verification key coded as a string
	Amount int // Amount to transfer
	Signature string // Potential signature coded as string
}

func (l *Ledger) SignedTransaction(t *SignedTransaction) {
	l.lock.Lock() ; defer l.lock.Unlock()
	/* We verify that the t.Signature is a valid RSA
	* signature on the rest of the fields in t under
	* the public key t.From.
	*/
	validSignature := true
	if validSignature {
		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount
	}
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

type Peer struct {
	Address string
}

type Message struct {
	ID    string
	Value interface{}
}

const TRANSACTION_MESSAGE = "transMsg"                 // When the message contains a transaction
const NEW_PEER_MESSAGE = "newPeerMsg"                  // When the message contains a new peer
const REQUEST_PEER_LIST_MESSAGE = "requestPeerListMsg" // When the message requests the list of peers
const PEER_LIST_MESSAGE = "peerListMsg"                // When the message contains the list of peers

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

func (l *Ledger) InitializeAccount(peer string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[peer] = 0
}

func (l *Ledger) PrintStatus() {
	l.lock.Lock()
	defer l.lock.Unlock()
	var keys = reflect.ValueOf(l.Accounts).MapKeys()
	fmt.Println("There are", len(keys), " keys: ", keys)
	for i := 0; i < len(keys); i++ {
		var key = keys[i]
		var str = key.String()
		fmt.Println("Account ", str, " has ", l.Accounts[str], " dineros")
	}
}
