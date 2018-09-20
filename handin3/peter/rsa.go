package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
)

type Key struct {
	n *big.Int
	e *big.Int
	d *big.Int
}

var e = big.NewInt(3)

func CalculatePrime(len int) *big.Int {
	var res *big.Int

	for {
		prime, err := rand.Prime(rand.Reader, len)
		if err != nil {
			fmt.Println("Error while calculating prime:", err.Error())
			return nil
		}

		// If (prime - 1) is divisible by e, find a new one
		if isGCD1(prime) {
			res = prime
			break
		}
	}

	return res
}

func KeyGen(k int) *Key {

	fmt.Println("Generating new RSA key")

	if k%2 != 0 {
		fmt.Println("Error: k needs to be an even number")
		return nil
	}

	// The bit length of p and q
	var primeLength = k / 2
	var p, q *big.Int

	p = CalculatePrime(primeLength)
	q = CalculatePrime(primeLength)

	fmt.Println("p:", p)
	fmt.Println("q:", q)

	// Calculate n = p * q
	var n = new(big.Int)
	n.Mul(p, q)
	fmt.Println("n:", n)

	var mult = new(big.Int).Mul(subtractInt(p, 1), subtractInt(q, 1))
	var d = new(big.Int).ModInverse(e, mult)

	fmt.Println("d:", d)
	fmt.Println("e:", e)

	fmt.Println("Public key (e, n): (" + e.String() + ", " + n.String() + ")")
	fmt.Println("Secret key (d, n): (" + d.String() + ", " + n.String() + ")")

	var res = Key{n: n, e: e, d: d}

	return &res
}

// For a prime p this method calculates: gcd(e, (p - 1)) = 1
func isGCD1(prime *big.Int) bool {
	var minusOne = subtractInt(prime, 1)

	// Calculate GCD
	var temp = new(big.Int)
	temp.GCD(nil, nil, e, minusOne)

	// Compare result to 1, return 0 if temp == 1
	var result = temp.Cmp(big.NewInt(1))

	return (result == 0)
}

// Subtracts a normal integer from a big.Int, and returns a new big.Int containing the result
func subtractInt(bigI *big.Int, i int) *big.Int {
	var res big.Int
	res.Sub(bigI, big.NewInt(int64(i)))
	return &res
}

func Encrypt(e, n, message *big.Int) *big.Int {
	var pow = new(big.Int)
	pow.Exp(message, e, nil)
	var res = new(big.Int)
	res.Mod(pow, n)
	return res
}

func Decrypt(d, n, ciphertext *big.Int) *big.Int {
	var pow = new(big.Int).Exp(ciphertext, d, nil)
	var res = new(big.Int)
	res.Mod(pow, n)
	return res
}

func GenerateNonce(length int) []byte {
	nonce := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	return nonce
}

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

func DecryptFromFile(key, nonce []byte, fileName string) {

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
}

func testRSA() {
	var key = KeyGen(16)

	// message < n
	var message = big.NewInt(1337)

	fmt.Println("Message:", message)

	var ciphertext = Encrypt(key.e, key.n, message)
	fmt.Println("Ciphertext:", ciphertext)

	var decryptedMessage = Decrypt(key.d, key.n, ciphertext)
	fmt.Println("Decrypted message:", decryptedMessage)
}

func testAES() {
	var fileName = "missionDetails"

	var message = []byte("Super secret message")
	var key = []byte("secretkeyteksjel")
	var nonce = EncryptToFile(key, message, fileName)

	DecryptFromFile(key, nonce, fileName)
}

func main() {

	testRSA()

	testAES()
}
