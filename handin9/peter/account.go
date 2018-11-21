package main

import (
	"math/big"
)

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

type InitInfo struct {
	Peers        []Peer
	GenesisBlock GenesisBlock
}

const TRANSACTION_MESSAGE = "transMsg"                 // When the message contains a transaction
const NEW_PEER_MESSAGE = "newPeerMsg"                  // When the message contains a new peer
const REQUEST_INIT_INFO_MESSAGE = "requestInitInfoMsg" // When the message requests the initial info
const INIT_INFO_MESSAGE = "initInfoMsg"                // When the message contains the initial info
const BLOCK_MESSAGE = "blockMsg"                       // When a block is sent

func (t *SignedTransaction) isValid() bool {
	// Get public key of the sender

	if t.Amount < 1 {
		return false
	}

	pk := GeneratePublicKeyFromString(t.From)

	signatureString := t.Signature
	signature := new(big.Int)
	signature.SetString(signatureString, 10)
	message := GenerateMessageFromTransaction(t)

	verify := Verify(message, signature, pk)

	return verify
}
