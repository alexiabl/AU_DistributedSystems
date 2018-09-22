package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
)

func ultimateTest() {
	// Generate key
	var keyRSA = KeyGen(16)
	fmt.Println("Key:", keyRSA)

	// Encrypt a message with the key
	var message = big.NewInt(1337)
	var ciphertext = Encrypt(keyRSA.E, keyRSA.N, message)
	fmt.Println("Message:", message)
	fmt.Println("Ciphertext:", ciphertext)

	// Convert key to byte array
	var bytes = ConvertKeyToByteArray(*keyRSA)

	// Save the key to a file
	var keyAES = GenerateNonce(16)
	var fileName = "EncryptedKey"
	var nonce = EncryptToFile(keyAES, bytes, fileName)

	// Load the key from a file
	var encodedKey = DecryptFromFile(keyAES, nonce, fileName)
	var decodedKey = ConvertByteArrayToKey(encodedKey)
	fmt.Println("Decoded key:", decodedKey)

	// Decrypt the message with the key
	var decryptedMessage = Decrypt(decodedKey.D, decodedKey.N, ciphertext)
	fmt.Println("Decrypted message:", decryptedMessage)
	var cmp = decryptedMessage.Cmp(message)

	// If they're equal: Happiness
	if cmp == 0 {
		fmt.Println("Happiness")
	} else {
		fmt.Println("Sadness")
	}

	// Remove the file
	os.Remove(fileName)
}

func ConvertKeyToByteArray(key Key) []byte {
	var buffer bytes.Buffer
	var writer = bufio.NewWriter(&buffer)
	var reader = bufio.NewReader(&buffer)

	// The bytes to return
	var bytes []byte

	var enc = gob.NewEncoder(writer)
	var err = enc.Encode(key)
	if err != nil {
		panic("Error while encoding key: " + err.Error())
	} else {
		writer.Flush()

		bytes = make([]byte, buffer.Len())
		_, err := reader.Read(bytes)

		if err != nil {
			panic("Error writing bytes to byte array: " + err.Error())
		}
	}

	return bytes
}

func ConvertByteArrayToKey(data []byte) *Key {
	var buffer bytes.Buffer
	buffer.Write(data)
	var reader = bufio.NewReader(&buffer)

	var dec = gob.NewDecoder(reader)
	var newKey = new(Key)
	err := dec.Decode(newKey)

	if err != nil {
		panic("Unable to decode the key: " + err.Error())
	}

	return newKey
}

func main() {
	ultimateTest()
}
