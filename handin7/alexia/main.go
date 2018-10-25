package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"time"
)

func basicVerification() {

	fmt.Println(" ----- Showcasing that signing and verification works -----")

	// Generate a RSA key
	n, d := KeyGen(256)
	pk := generatePublicKey(n, e)
	sk := generateSecretKey(n, d)

	// Sign a message
	messageString := "When I wake up, the other side of the bed is cold. My fingers stretch out, seeking Prim’s warmth but finding only the rough canvas cover of the mattress. She must have had bad dreams and climbed in with our mother. Of course, she did. This is the day of the reaping."
	message := []byte(messageString)
	signature := sign(message, sk)

	// Print the message as a string and a byte array
	fmt.Println("Message (string):", messageString, "\n")
	fmt.Println("Message (byte array):", message, "\n")

	// Print the generated signature
	fmt.Println("Signature:", signature)

	// Verifying a valid message
	fmt.Println("Verifying valid message:", verify(message, signature, pk))

	// Verifying a fake message
	fakeMessage := []byte("When I wake up")
	fmt.Println("Verifying fake message:", verify(fakeMessage, signature, pk))
}

func testHashSpeed() {

	fmt.Println(" ----- Testing sha256 hash speed on 10kb file -----")

	filename := "text/large_file.txt"
	message, err := ioutil.ReadFile(filename)

	bits := len(message) * 8

	if err != nil {
		panic(err.Error())
	}

	// Setup sha
	sha := sha256.New()

	// Setup timer
	timeStart := time.Now()

	// Do actual hashing
	sha.Write(message)

	// Get time passed
	endTime := time.Since(timeStart)
	seconds := endTime.Seconds()

	// Calculate bits per second
	bitsPerSecond := float64(bits) / seconds

	fmt.Println("Bits hashed:", bits)
	fmt.Println("Hashing time:", endTime)
	fmt.Printf("Hashing time (seconds): %f\n", seconds)
	fmt.Printf("Bits per second: %f\n", bitsPerSecond)
}

func testSignatureSpeed() {

	fmt.Println(" ----- Testing how long it takes to sign with a 2000 bit key -----")

	// Generate a 2000 bit key
	n, d := KeyGen(2000)
	sk := generateSecretKey(n, d)

	// Setup the message to be signed
	message := make([]byte, 1999)

	// Calculate average of 100 runs
	var sum = time.Duration(0)
	for i := 0; i < 100; i++ {
		// Setup start time
		startTime := time.Now()

		// Do the actual signing
		sign(message, sk)

		// Calculate how long it took
		endTime := time.Since(startTime)
		sum += endTime
	}

	average := sum / 100

	fmt.Println("Signing time (average):", average)
}

func testSoftwareWallet(){
	pwd := GetPassword()
	Generate("softwareWalletTest",pwd)
	message := "This is a message"
	pwd = GetPassword()
	signature := Sign("softwareWalletTest", pwd,[]byte(message))
	fmt.Println(signature)
}

func testGenerate() string {
	pwd := GetPassword()
	fmt.Print("Enter the filename: ")
	var filename string
	fmt.Scan(&filename)
	pk := Generate("text/"+filename,pwd)
	return pk
}

func testSign (pk string){
	msg := "This is a test message"
	pwd := GetPassword()
	fmt.Print("Enter the filename: ")
	var filename string
	fmt.Scan(&filename)
	signature := Sign("text/"+filename, pwd,[]byte(msg))
	public_key := GeneratePublicKeyFromString(pk)
	signature_ok := verify([]byte(msg),signature,public_key)
	fmt.Println("Signature OK: ",signature_ok)
}

func main() {
	pk := testGenerate()
	testSign(pk)
}
