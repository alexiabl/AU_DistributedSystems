package main

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"strconv"
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

type Block struct {
	ID	int
	Transactions []string
	Signature	 *big.Int
}

type Sequencer struct {
	Pk		string
	Sk		string
	Ip 		string
}

const TRANSACTION_MESSAGE = "transMsg"                 // When the message contains a transaction
const NEW_PEER_MESSAGE = "newPeerMsg"                  // When the message contains a new peer
const REQUEST_PEER_LIST_MESSAGE = "requestPeerListMsg" // When the message requests the list of peers
const PEER_LIST_MESSAGE = "peerListMsg"                // When the message contains the list of peers
const BLOCK_MESSAGE = "blockMsg"							   // When a block is sent
const REQUEST_SEQUENCER_MESSAGE = "requestSeqMsg"				   // Ask who is the sequencer
const SEQUENCER_MESSAGE = "SeqMsg"							// Contains the sequencer key

func (l *Ledger) SignedTransaction(t *SignedTransaction) bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Get public key of the sender
	pk := GeneratePublicKeyFromString(t.From)

	signatureString := t.Signature
	signature := new(big.Int)
	signature.SetString(signatureString, 10)
	message := GenerateMessageFromTransaction(t)

	validSignature := Verify(message, signature, pk)

	if validSignature {
		if l.Accounts[t.From] - t.Amount > 0 {
			l.Accounts[t.From] -= t.Amount
			l.Accounts[t.To] += t.Amount
			return true
		}
	}

	return false
}

func GenerateMessageFromTransaction(t *SignedTransaction) []byte {
	return []byte(t.From + t.To + t.ID + string(t.Amount))
}

func GenerateMessageFromBlock(block Block) ([]byte,error) {
	idBytes := []byte(strconv.Itoa(block.ID))
	for i := 0 ; i<len(block.Transactions); i++ {
		idBytes = append(idBytes,[]byte(block.Transactions[i])...)
	}
    return idBytes, nil
}

func (l *Ledger) InitializeAccount(peer Peer) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[peer.Pk] = 1000
}

func (l *Ledger) PrintStatus() {
	l.lock.Lock()
	defer l.lock.Unlock()
	var keys = reflect.ValueOf(l.Accounts).MapKeys()
	fmt.Println("There are", len(keys), "keys")
	for i := 0; i < len(keys); i++ {
		var key = keys[i]
		var str = key.String()
		var peer = GetPeerFromPK(str)
		fmt.Println("Account", peer.Address, "has", l.Accounts[str], "dineros")
	}
}

func GetPeerFromPK(str string) *Peer {
	for i := 0; i < len(peers); i++ {
		if str == peers[i].Pk {
			return &peers[i]
		}
	}

	return nil
}

func GetPeerFromIP(ip string) *Peer {
	for i := 0; i < len(peers); i++ {
		if ip == peers[i].Address {
			return &peers[i]
		}
	}

	return nil
}
