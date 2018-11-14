package main

import "math/big"

func (l *Ledger) SignedTransaction(t *SignedTransaction) bool {
	l.Lock.Lock()
	defer l.Lock.Unlock()

	// Get public key of the sender
	pk := GeneratePublicKeyFromString(t.From)

	signatureString := t.Signature
	signature := new(big.Int)
	signature.SetString(signatureString, 10)
	message := GenerateMessageFromTransaction(t)

	validSignature := Verify(message, signature, pk)

	if validSignature {
		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount
		return true
	}

	return false
}

func GenerateMessageFromTransaction(t *SignedTransaction) []byte {
	return []byte(t.From + t.To + t.ID + string(t.Amount))
}
