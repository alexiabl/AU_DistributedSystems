package main

import (
	"strconv"
)

func GenerateMessageFromTransaction(t *SignedTransaction) []byte {
	return []byte(t.From + t.To + t.ID + string(t.Amount))
}

func GenerateMessageFromBlock(block *Block) []byte {
	idBytes := []byte(strconv.Itoa(block.ID) + block.PreviousBlock + block.Sender)

	for i := 0; i < len(block.Transactions); i++ {
		idBytes = append(idBytes, []byte(block.Transactions[i])...)
	}

	return idBytes
}
