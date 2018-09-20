package main

import (
	"fmt"
	"math/big"
	"crypto/rand"
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
)

var e *big.Int = big.NewInt(3)
var filename = "aes.txt"

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

// RSA Decryption method
func Decrypt(ciphertext *big.Int, n *big.Int, p *big.Int, q *big.Int){
	//d = 3^-1 mod (p-1)(q-1)
	d := new(big.Int)
	mult := new(big.Int)
	mult.Mul(subtract(p,1),subtract(q,1))
	d.ModInverse(e,mult)
	fmt.Println("d = ",d)

	temp := new(big.Int)
	message := new(big.Int)
	message.Mod(temp.Exp(ciphertext,d,nil),n)
	fmt.Println("message = ", message)
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
	fmt.Println("decrypted message from file: ",string(message[:]))
	return message

}

// Method to test RSA individually
func testRSA(){
	n,p,q:=KeyGen(21)
	fmt.Println("p = ",p)
	fmt.Println("q = ",q)
	original_msg := big.NewInt(11215)
	fmt.Println("original message = ",original_msg)
	cipher := Encrypt(original_msg,n)
	Decrypt(cipher,n,p,q)
}

// Method to test AES individually
func testAES(){
	key := []byte("1098765432100000")
	message := []byte("ThisIsAnEncrypedMessage")
	fmt.Println("original message: ",string(message[:]))
	EncryptToFile(message,key)
	DecryptFromFile(key)
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
	n,_,_:=KeyGen(21)
	EncryptToFile(n.Bytes(),key)
	DecryptFromFile(key)


}
