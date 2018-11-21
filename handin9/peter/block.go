package main

import "math/big"

type Block struct {
	ID            int
	PreviousBlock string
	Sender        string
	Transactions  []string
	Signature     string
	Draw          *big.Int
}

type GenesisBlock struct {
	*Block
	KingKeys []string
	Seed     int
}

func (b *Block) isValid() bool {

	pk := GeneratePublicKeyFromString(b.Sender)

	blockMsg := GenerateMessageFromBlock(b)

	temp := new(big.Int)
	signature, _ := temp.SetString(b.Signature, 10)

	return Verify(blockMsg, signature, pk)
}
