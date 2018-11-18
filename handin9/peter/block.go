package main

import "math/big"

type Block struct {
	ID            int
	PreviousBlock string
	Sender        string
	Transactions  []string
	Signature     string
}

type GenesisBlock struct {
	*Block
	KingKeys []string
	Seed     int
}

func (b *Block) isValidSignature() bool {
	blockMsg := GenerateMessageFromBlock(b)

	temp := new(big.Int)
	signature, _ := temp.SetString(b.Signature, 10)

	pk := GeneratePublicKeyFromString(b.Sender)

	return Verify(blockMsg, signature, pk)
}
