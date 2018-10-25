package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
)

// Create a nonce for AES encryption and decryption
func CreateNonce(len int) []byte {
	nonce := make([]byte, len)
	return nonce
}

// AES encryption that writes the ciphertext to a file
func EncryptToFile(message []byte, key []byte, filename string) []byte {
	block, err := aes.NewCipher(key)
	nonce := CreateNonce(12)
	if err != nil {
		panic(err.Error())
	}
	gcm, _ := cipher.NewGCM(block)

	ciphertext := gcm.Seal(nil, nonce, message, nil)
	ioutil.WriteFile(filename, ciphertext, 0644)
	fmt.Println("AES encryption ciphertext: ", ciphertext)
	return ciphertext
}

// AES decryption from the file
func DecryptFromFile(key []byte, filename string) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	ciphertext, err := ioutil.ReadFile(filename)
	nonce := CreateNonce(12)
	if err != nil {
		panic(err.Error())
	}
	gcm, _ := cipher.NewGCM(block)
	message, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return message

}
