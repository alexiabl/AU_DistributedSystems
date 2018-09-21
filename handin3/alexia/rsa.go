package main

import (
	"fmt"
	"math/big"
	"crypto/rand"
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
	"encoding/gob"
	"bytes"
	"bufio"
)

var e *big.Int = big.NewInt(3)
var filename = "aes.txt"

type PublicKey struct {
	N_pk *big.Int
	E_pk *big.Int
}

type SecretKey struct {
	N_sk *big.Int
	D_sk *big.Int
}

//RSA Key generator
func KeyGen(k int) (*big.Int, *big.Int, *big.Int){
	//check if it is even
	var p,q *big.Int
	n := new(big.Int)

		p = calculatePrime(k)
		q = calculatePrime(k)
		n.Mul(p,q)

	//var n  = p * q
	fmt.Println("bit length n =",n.BitLen())
	return n,p,q
}

func generatePublicKey(n *big.Int, e *big.Int) (PublicKey){
	pk := new(PublicKey)
	pk.N_pk = n
	pk.E_pk = e

	return *pk
}

// Helper method to calculate a prime number and check the GCD condition
func calculatePrime(k int) (*big.Int){
	for{
		prime, err := rand.Prime(rand.Reader,int(k/2))
		if (err != nil){
			fmt.Println("Error generating prime ",err)
		}
		if (TestGCD(prime)){
			return prime
		}
	}
}

// RSA Encryption method
func Encrypt(message *big.Int, n *big.Int) (*big.Int){
	cipher := new(big.Int)
	exp := cipher.Exp(message,e,nil)
	cipher.Mod(exp,n)
	fmt.Println("ciphertext = ",cipher)
	return cipher
}

// Test the GCD condition for the prime numbers
func TestGCD(prime *big.Int) bool{
	sub := new(big.Int).Sub(prime,big.NewInt(1))
	if (new(big.Int).GCD(nil,nil,e,sub).Cmp(big.NewInt(1))==0){
		return true
	}else{
		return false
	}
}

// Helper method to subtract from big.Int
func subtract(prime *big.Int, i int64) (result *big.Int){
	result = new(big.Int)
	one := big.NewInt(i)
	result.Sub(prime,one)
	return result
}

func generateSecretKey(n *big.Int, d *big.Int)(SecretKey){
	sk := new(SecretKey)
	sk.N_sk = n
	sk.D_sk = d
	return *sk
}

// Helper method to calculate D
func calculateD(p *big.Int, q *big.Int)(*big.Int){
	d := new(big.Int)
	mult := new(big.Int)
	mult.Mul(subtract(p,1),subtract(q,1))
	d.ModInverse(e,mult)
	return d
}

// RSA Decryption method
func Decrypt(ciphertext *big.Int, n *big.Int, p *big.Int, q *big.Int)(*big.Int, *big.Int){
	//d = 3^-1 mod (p-1)(q-1)
	d := calculateD(p,q)
	temp := new(big.Int)
	message := new(big.Int)
	message.Mod(temp.Exp(ciphertext,d,nil),n)
	return d,message
}

// Create a nonce for AES encryption and decryption
func CreateNonce(len int) ([]byte) {
	nonce := make([]byte,len)
	return nonce
}

// AES encryption that writes the ciphertext to a file
func EncryptToFile(message []byte, key []byte) ([]byte) {
	block, err := aes.NewCipher(key)
	nonce := CreateNonce(12)
	if err != nil {
		panic(err.Error())
	}
	gcm, _ := cipher.NewGCM(block)

	ciphertext := gcm.Seal(nil,nonce,message,nil)
	ioutil.WriteFile(filename,ciphertext,0644)
	fmt.Println("ciphertext aes: ",ciphertext)
	return ciphertext
}

// AES decryption from the file 
func DecryptFromFile(key []byte) ([]byte){
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
	//fmt.Println("decrypted message from file: ",string(message[:]))
	return message

}

// Method to test RSA individually
func testRSA(){
	n,p,q:=KeyGen(21)
	fmt.Println("p = ",p)
	fmt.Println("q = ",q)
	fmt.Println("Public Key (n,e)",n,e)
	original_msg := big.NewInt(13215)
	fmt.Println("original message = ",original_msg)
	cipher := Encrypt(original_msg,n)
	d,message := Decrypt(cipher,n,p,q)
	fmt.Println("d = ",d)
	fmt.Println("decrypted message = ",message)
}

// Method to test AES individually
func testAES(){
	key := []byte("1098765432100000")
	message := []byte("Hi my name is Alexia")
	fmt.Println("original message: ",string(message[:]))
	EncryptToFile(message,key)
	decrypted := DecryptFromFile(key)
	fmt.Println("decrypted message from file: ",string(decrypted[:]))
}



func main(){
	fmt.Println("RSA TEST:")
	fmt.Println("=============")
	testRSA()

	fmt.Println("\nAES TEST:")
	fmt.Println("=============")
	testAES()

	fmt.Println("\nTESTING RSA WITH AES:")
	fmt.Println("=============")
	key := []byte("1098765432100000")
	n,p,q:=KeyGen(14)

	//key in RSA (n,e)
	pk := generatePublicKey(n,e)
	fmt.Println("Public Key (n,e): ",pk.N_pk, pk.E_pk)
	d := calculateD(p,q)
	sk := generateSecretKey(n,d)
	fmt.Println("Secret Key (n,d): ",sk.N_sk,sk.D_sk)
	
	var buffer bytes.Buffer
	var writer = bufio.NewWriter(&buffer)	
	var encoder = gob.NewEncoder(writer)
	err := encoder.Encode(&pk)
	if (err != nil){
		panic(err.Error())
	} else{
		writer.Flush()
	}
	fmt.Println("Encoded public key: ",buffer)
	EncryptToFile(buffer.Bytes(),key)
	encoded_key := DecryptFromFile(key)
	fmt.Println("Decrypted encoded key: ",encoded_key)

	buffer.Reset()
	buffer.Write(encoded_key)
	var reader = bufio.NewReader(&buffer)
	var publicKey = &PublicKey{}
	var dec = gob.NewDecoder(reader)
	err = dec.Decode(publicKey)
	if (err != nil){
		fmt.Println("Error - ", err.Error())
	}
	fmt.Println("Decrypted key: ", publicKey.N_pk,publicKey.E_pk)


}
