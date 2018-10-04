package main

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
)

type Ledger struct {
	Accounts map[string]int
	lock     sync.Mutex
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

func (l *Ledger) SignedTransaction(t *SignedTransaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Get public key of the sender
	pk := new(PublicKey)
	pk.fromString(t.From)

	signatureString := t.Signature
	signature := new(big.Int)
	signature.SetString(signatureString, 10)
	message := GenerateMessageFromTransaction(t)

	validSignature := verify(message, signature, *pk)

	/* We verify that the t.Signature is a valid RSA
	 * signature on the rest of the fields in t under
	 * the public key t.From.
	 */

	if validSignature {
		l.Accounts[t.From] -= t.Amount
	}
}

func GenerateMessageFromTransaction(t *SignedTransaction) []byte {
	return []byte(t.From + t.To + t.ID + string(t.Amount))
}

func (l *Ledger) InitializeAccount(peer Peer) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[peer.Pk] = 0
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
