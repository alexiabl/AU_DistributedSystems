package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
)

// Encrypts a message to a file given a key, and returns the nonce
func EncryptToFile(key, message []byte, fileName string) []byte {

	fmt.Println("------------------------------")
	fmt.Println("Encrypting a message to the file:", fileName)
	fmt.Println("Message:", string(message))
	fmt.Println("Key:", string(key))

	var nonce = GenerateNonce(12)
	fmt.Println("Nonce:", string(nonce))

	block, err := aes.NewCipher(key)

	if err != nil {
		panic("Error while creating AES Cipher: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		panic("Error while creating AES GCM: " + err.Error())
	}

	var msg = aesgcm.Seal(nil, nonce, message, nil)

	fmt.Println("Encrypted message:", string(msg))

	err = ioutil.WriteFile(fileName, msg, 0644) //TODO: Get mode by code

	if err != nil {
		panic("Unable to open file:" + err.Error())
	}

	fmt.Println(" ----- File successfully written -----")

	return nonce
}

func DecryptFromFile(key, nonce []byte, fileName string) []byte {

	fmt.Println("------------------------------")
	fmt.Println("Decrypting the file:", fileName)
	fmt.Println("Key:", string(key))
	fmt.Println("Nonce:", string(nonce))

	file, err := ioutil.ReadFile(fileName)

	if err != nil {
		panic("Unable to open file: " + err.Error())
	}

	var ciphertext = []byte(file)

	fmt.Println("Encrypted message:", string(ciphertext))

	block, err := aes.NewCipher(key)

	if err != nil {
		panic("Error while creating AES Cipher: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		panic("Error while creating AES GCM: " + err.Error())
	}

	msg, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		panic("Error decrypting the ciphertext: " + err.Error())
	}

	fmt.Println("Decrypted message:", string(msg))
	fmt.Println(" ----- Decryption successful -----")

	return msg
}
