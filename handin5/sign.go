package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

func sign(message []byte, sk SecretKey) *big.Int {

	// Signing:
	// s = S(m) = m ^ d mod n

	// Hash the message
	sha := sha256.New()
	sha.Write(message)
	hash := sha.Sum(nil)

	// Convert hash to big int
	messageInt := big.NewInt(0).SetBytes(hash)

	// Sign the int
	signedInt := new(big.Int).Exp(messageInt, sk.D_sk, sk.N_sk)

	return signedInt
}

func verify(message []byte, signature *big.Int, pk PublicKey) bool {

	// Verifying:
	// m = s ^ e mod n

	// 'Decrypt' the signature
	decSign := new(big.Int).Exp(signature, pk.E_pk, pk.N_pk)
	decSignBytes := decSign.Bytes()

	// Calculate the sha of the original message
	sha := sha256.New()
	sha.Write(message)
	hash := sha.Sum(nil)

	// Return whether the decrypted signature matches the sha of the message
	return (bytes.Compare(hash, decSignBytes) == 0)
}
