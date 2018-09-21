package main

import (
	"fmt"
	"math/big"
	"crypto/rand"
)

var e *big.Int = big.NewInt(3)

type PublicKey struct {
	N_pk *big.Int
	E_pk *big.Int
}

type SecretKey struct {
	N_sk *big.Int
	D_sk *big.Int
}

//RSA Key generator
func KeyGen(k int) (*big.Int, *big.Int){
	//check if it is even
	var p,q *big.Int
	n := new(big.Int)

		p = calculatePrime(k)
		q = calculatePrime(k)
		n.Mul(p,q)

	//var n  = p * q
	d := calculateD(p,q)
	fmt.Println("bit length n =",n.BitLen())
	return n,d
}

// Creates and returns a public key object with the n and e sent
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
	//fmt.Println("ciphertext = ",cipher)
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

// Creates and returns Secret key object with the n and d sent
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
func Decrypt(ciphertext *big.Int, n *big.Int, d *big.Int)(*big.Int, *big.Int){
	//d = 3^-1 mod (p-1)(q-1)
	temp := new(big.Int)
	message := new(big.Int)
	message.Mod(temp.Exp(ciphertext,d,nil),n)
	return d,message
}

// Method to test RSA individually
func testRSA(){
	n,d:=KeyGen(21)
	fmt.Println("Public Key (n,e)",n,e)
	original_msg := big.NewInt(13215)
	fmt.Println("original message = ",original_msg)
	cipher := Encrypt(original_msg,n)
	fmt.Println("ciphertext = ",cipher)
	d,message := Decrypt(cipher,n,d)
	fmt.Println("d = ",d)
	fmt.Println("decrypted message = ",message)
}

