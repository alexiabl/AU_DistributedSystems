package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"math/big"
	"golang.org/x/crypto/bcrypt"
)

var my_filename = ""

// Create a nonce for AES encryption and decryption
func CreateNonce(len int) []byte {
	nonce := make([]byte, len)
	return nonce
}

//Ask for password, salt -> hash

func GetPassword() string {
	// Prompt the user to enter a password
    fmt.Print("Enter a password (16 characters): ")
    // Variable to store the users input
    var pwd string
    // Read the users input
    _, err := fmt.Scan(&pwd)
    if err != nil {
        fmt.Println(err)
    }
    return pwd
}

// Returns hash from password entered
func HashSalt(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
    if err != nil {
        fmt.Println(err)
    }
    return string(hash)
}

// Generates RSA public and secret keys and encrypts the secret key to the file specified using the password 
func Generate(filename string, password string) string {
	n, d := KeyGen(256)
	p_k := generatePublicKey(n,e)
	s_k := generateSecretKey(n,d)
	//Salt and hash password for security measures, we will compare this hash later when we try to decrypt the secret key from the file
	hash := HashSalt(password)
	//Save hash in file
	ioutil.WriteFile(filename+"_hash", []byte(hash), 0644)
	my_filename = filename
	EncryptToFile([]byte(s_k.toString()),[]byte(password),filename)
	return p_k.toString()
}

// Verifies if hashed password from file matches the input password
func VerifyPassword(hashedPwd []byte, inputPassword []byte) bool{
    err := bcrypt.CompareHashAndPassword(hashedPwd, inputPassword)
    if err != nil {
        fmt.Println(err)
        return false
    }
    return true
}

func Sign(filename string, password string, msg []byte) *big.Int{
	hashPwd_fromFile, _ := ioutil.ReadFile(filename+"_hash")
	var signature = new(big.Int)
	if my_filename == filename && VerifyPassword(hashPwd_fromFile, []byte(password)) {
		sk_byte := DecryptFromFile([]byte(password),filename)
		sk_str := string(sk_byte)
		secret_key := GenerateSecretKeyFromString(sk_str)
		signature = sign(msg, secret_key)
		fmt.Println("Signature = "+signature.String())
	} else {
		fmt.Println("ERROR: Access denied - incorrect password or mismatched filenames")
	}
	return signature
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



